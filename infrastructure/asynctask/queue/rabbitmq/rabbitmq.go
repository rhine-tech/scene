package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

const defaultExchange = "scene.asynctask"

type Config struct {
	URL          string
	Exchange     string
	ExchangeType string
	Durable      bool
	AutoDelete   bool
	Prefetch     int
}

type handlerEntry struct {
	handler asynctask.TaskQueueHandler
}

type consumerRuntime struct {
	config  asynctask.TaskQueueConfig
	started bool
}

type Queue struct {
	config    Config
	lock      sync.Mutex
	conn      *amqp.Connection
	handlers  map[string]map[string]handlerEntry
	consumers map[string]*consumerRuntime
}

func New(config Config) *Queue {
	if config.Exchange == "" {
		config.Exchange = defaultExchange
	}
	if config.ExchangeType == "" {
		config.ExchangeType = "direct"
	}
	if config.Prefetch <= 0 {
		config.Prefetch = 16
	}
	return &Queue{
		config:    config,
		handlers:  make(map[string]map[string]handlerEntry),
		consumers: make(map[string]*consumerRuntime),
	}
}

func (q *Queue) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskQueue", "rabbitmq")
}

func (q *Queue) Setup() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if err := q.ensureConnectedLocked(); err != nil {
		return err
	}
	for name, runtime := range q.consumers {
		if runtime.started {
			continue
		}
		if err := q.startConsumerLocked(name, runtime); err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) Dispose() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.conn == nil {
		return nil
	}
	err := q.conn.Close()
	q.conn = nil
	return err
}

func (q *Queue) Publish(ctx context.Context, task *asynctask.QueueTask) (*asynctask.QueueTask, error) {
	if task == nil || task.Queue == "" || task.Type == "" {
		return nil, asynctask.ErrInvalidQueueTask
	}
	task.Identifier()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.AvailableAt = task.CreatedAt.Add(task.Delay)
	task.Status = asynctask.QueueTaskStatusPending

	q.lock.Lock()
	err := q.ensureConnectedLocked()
	q.lock.Unlock()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	ch, err := q.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()
	if err := q.declareQueue(ch, task.Queue); err != nil {
		return nil, err
	}

	headers := amqp.Table{}
	for k, v := range task.Headers {
		headers[k] = v
	}
	headers["x-task-type"] = task.Type
	headers["x-task-attempt"] = int32(task.Attempt)

	if err := ch.PublishWithContext(ctx, q.config.Exchange, task.Queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
		Timestamp:   task.CreatedAt,
		Headers:     headers,
	}); err != nil {
		return nil, err
	}
	return task, nil
}

func (q *Queue) RegisterQueue(queue string, config asynctask.TaskQueueConfig) error {
	if queue == "" {
		return asynctask.ErrInvalidQueueName
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	normalized := normalizeConfig(config)
	runtime, ok := q.consumers[queue]
	if ok && !sameConfig(runtime.config, normalized) {
		return asynctask.ErrQueueConfigConflict.WithDetailStr(queue)
	}
	if !ok {
		runtime = &consumerRuntime{config: normalized}
		q.consumers[queue] = runtime
	}
	if q.conn != nil && !runtime.started {
		return q.startConsumerLocked(queue, runtime)
	}
	return nil
}

func (q *Queue) RegisterHandler(queue string, taskType string, handler asynctask.TaskQueueHandler) error {
	if queue == "" {
		return asynctask.ErrInvalidQueueName
	}
	if taskType == "" {
		return asynctask.ErrInvalidTaskType
	}
	if handler == nil {
		return asynctask.ErrInvalidTaskHandler
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	if _, ok := q.consumers[queue]; !ok {
		return asynctask.ErrQueueNotRegistered.WithDetailStr(queue)
	}
	queueHandlers, ok := q.handlers[queue]
	if !ok {
		queueHandlers = make(map[string]handlerEntry)
		q.handlers[queue] = queueHandlers
	}
	queueHandlers[taskType] = handlerEntry{handler: handler}
	return nil
}

func (q *Queue) ensureConnectedLocked() error {
	if q.conn != nil && !q.conn.IsClosed() {
		return nil
	}
	if q.config.URL == "" {
		return fmt.Errorf("rabbitmq url is empty")
	}
	conn, err := amqp.Dial(q.config.URL)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(
		q.config.Exchange,
		q.config.ExchangeType,
		q.config.Durable,
		q.config.AutoDelete,
		false,
		false,
		nil,
	); err != nil {
		_ = conn.Close()
		return err
	}
	q.conn = conn
	return nil
}

func (q *Queue) startConsumerLocked(queueName string, runtime *consumerRuntime) error {
	if err := q.ensureConnectedLocked(); err != nil {
		return err
	}
	ch, err := q.conn.Channel()
	if err != nil {
		return err
	}
	if err := q.declareQueue(ch, queueName); err != nil {
		_ = ch.Close()
		return err
	}
	prefetch := q.config.Prefetch
	if runtime.config.Concurrency > prefetch {
		prefetch = runtime.config.Concurrency
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		_ = ch.Close()
		return err
	}
	deliveries, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return err
	}
	runtime.started = true
	for i := 0; i < runtime.config.Concurrency; i++ {
		go q.consume(queueName, runtime, deliveries)
	}
	return nil
}

func (q *Queue) declareQueue(ch *amqp.Channel, queueName string) error {
	if err := ch.ExchangeDeclare(
		q.config.Exchange,
		q.config.ExchangeType,
		q.config.Durable,
		q.config.AutoDelete,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queueName, q.config.Durable, q.config.AutoDelete, false, false, nil); err != nil {
		return err
	}
	return ch.QueueBind(queueName, queueName, q.config.Exchange, false, nil)
}

func (q *Queue) consume(queueName string, runtime *consumerRuntime, deliveries <-chan amqp.Delivery) {
	for delivery := range deliveries {
		task := &asynctask.QueueTask{}
		if err := json.Unmarshal(delivery.Body, task); err != nil {
			_ = delivery.Reject(false)
			continue
		}
		handler, err := q.lookupHandler(queueName, task.Type)
		if err != nil {
			_ = delivery.Reject(false)
			continue
		}
		task.Status = asynctask.QueueTaskStatusRunning
		ctx := context.Background()
		cancel := func() {}
		if task.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, task.Timeout)
		}
		handleErr := func() (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("rabbitmq task panic: %v", recovered)
				}
			}()
			return handler.HandleTask(ctx, task)
		}()
		cancel()
		if handleErr == nil {
			task.Status = asynctask.QueueTaskStatusSucceeded
			_ = delivery.Ack(false)
			continue
		}

		task.LastError = handleErr.Error()
		maxRetry := runtime.config.MaxRetry
		if task.MaxRetry > 0 {
			maxRetry = task.MaxRetry
		}
		if task.Attempt >= maxRetry {
			task.Status = asynctask.QueueTaskStatusFailed
			_ = delivery.Ack(false)
			continue
		}

		task.Attempt++
		task.Status = asynctask.QueueTaskStatusRetrying
		delay := runtime.config.RetryDelay
		if delay <= 0 {
			delay = time.Second
		}
		time.Sleep(delay)
		if _, err := q.Publish(context.Background(), task); err != nil {
			_ = delivery.Nack(false, true)
			continue
		}
		_ = delivery.Ack(false)
	}
}

func (q *Queue) lookupHandler(queueName string, taskType string) (asynctask.TaskQueueHandler, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	queueHandlers, ok := q.handlers[queueName]
	if !ok {
		return nil, asynctask.ErrTaskHandlerNotFound.WithDetailStr(queueName)
	}
	entry, ok := queueHandlers[taskType]
	if !ok {
		return nil, asynctask.ErrTaskHandlerNotFound.WithDetailStr(queueName + ":" + taskType)
	}
	return entry.handler, nil
}

func normalizeConfig(config asynctask.TaskQueueConfig) asynctask.TaskQueueConfig {
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second
	}
	return config
}

func sameConfig(left asynctask.TaskQueueConfig, right asynctask.TaskQueueConfig) bool {
	left = normalizeConfig(left)
	right = normalizeConfig(right)
	return left.Concurrency == right.Concurrency &&
		left.MaxRetry == right.MaxRetry &&
		left.RetryDelay == right.RetryDelay
}

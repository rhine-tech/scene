package redisstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/model"
	"github.com/spf13/cast"
)

const (
	defaultBlockDuration = 5 * time.Second
	defaultClaimIdle     = 30 * time.Second
	defaultPollInterval  = 5 * time.Second
	defaultStreamPrefix  = "scene.asynctask."
	defaultGroupPrefix   = "scene.asynctask."
)

type Config struct {
	Redis        model.DatabaseConfig
	StreamPrefix string
	GroupPrefix  string
	Block        time.Duration
	ClaimIdle    time.Duration
	PollInterval time.Duration
}

type handlerEntry struct {
	handler asynctask.TaskQueueHandler
}

type consumerRuntime struct {
	config       asynctask.TaskQueueConfig
	consumerName string
	started      bool
}

type Queue struct {
	config    Config
	lock      sync.RWMutex
	rdb       *redis.Client
	handlers  map[string]map[string]handlerEntry
	consumers map[string]*consumerRuntime
	stopCh    chan struct{}
}

func New(config Config) *Queue {
	if config.StreamPrefix == "" {
		config.StreamPrefix = defaultStreamPrefix
	}
	if config.GroupPrefix == "" {
		config.GroupPrefix = defaultGroupPrefix
	}
	if config.Block <= 0 {
		config.Block = defaultBlockDuration
	}
	if config.ClaimIdle <= 0 {
		config.ClaimIdle = defaultClaimIdle
	}
	if config.PollInterval <= 0 {
		config.PollInterval = defaultPollInterval
	}
	return &Queue{
		config:    config,
		handlers:  make(map[string]map[string]handlerEntry),
		consumers: make(map[string]*consumerRuntime),
		stopCh:    make(chan struct{}),
	}
}

func (q *Queue) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskQueue", "redisstream")
}

func (q *Queue) Setup() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if err := q.ensureClientLocked(); err != nil {
		return err
	}
	for queueName, runtime := range q.consumers {
		if runtime.started {
			continue
		}
		if err := q.startConsumerLocked(queueName, runtime); err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) Dispose() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	select {
	case <-q.stopCh:
	default:
		close(q.stopCh)
	}
	if q.rdb == nil {
		return nil
	}
	err := q.rdb.Close()
	q.rdb = nil
	return err
}

func (q *Queue) Publish(ctx context.Context, task *asynctask.QueueTask) (*asynctask.QueueTask, error) {
	if task == nil || strings.TrimSpace(task.Queue) == "" || strings.TrimSpace(task.Type) == "" {
		return nil, asynctask.ErrInvalidQueueTask
	}
	task.Identifier()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.AvailableAt = task.CreatedAt.Add(task.Delay)
	task.Status = asynctask.QueueTaskStatusPending

	payload, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	q.lock.Lock()
	err = q.ensureClientLocked()
	q.lock.Unlock()
	if err != nil {
		return nil, err
	}

	_, err = q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: q.streamName(task.Queue),
		Values: map[string]any{
			"task": payload,
		},
	}).Result()
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (q *Queue) RegisterQueue(queueName string, config asynctask.TaskQueueConfig) error {
	if strings.TrimSpace(queueName) == "" {
		return asynctask.ErrInvalidQueueName
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	normalized := normalizeConfig(config)
	runtime, ok := q.consumers[queueName]
	if ok && !sameConfig(runtime.config, normalized) {
		return asynctask.ErrQueueConfigConflict.WithDetailStr(queueName)
	}
	if !ok {
		runtime = &consumerRuntime{
			config:       normalized,
			consumerName: uuid.NewString(),
		}
		q.consumers[queueName] = runtime
	}
	if q.rdb != nil && !runtime.started {
		return q.startConsumerLocked(queueName, runtime)
	}
	return nil
}

func (q *Queue) RegisterHandler(queueName string, taskType string, handler asynctask.TaskQueueHandler) error {
	if strings.TrimSpace(queueName) == "" {
		return asynctask.ErrInvalidQueueName
	}
	if strings.TrimSpace(taskType) == "" {
		return asynctask.ErrInvalidTaskType
	}
	if handler == nil {
		return asynctask.ErrInvalidTaskHandler
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	if _, ok := q.consumers[queueName]; !ok {
		return asynctask.ErrQueueNotRegistered.WithDetailStr(queueName)
	}
	queueHandlers, ok := q.handlers[queueName]
	if !ok {
		queueHandlers = make(map[string]handlerEntry)
		q.handlers[queueName] = queueHandlers
	}
	queueHandlers[taskType] = handlerEntry{handler: handler}
	return nil
}

func (q *Queue) ensureClientLocked() error {
	if q.rdb != nil {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", q.config.Redis.Host, q.config.Redis.Port),
		Username: q.config.Redis.Username,
		Password: q.config.Redis.Password,
		DB:       cast.ToInt(q.config.Redis.Database),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		return err
	}
	q.rdb = rdb
	return nil
}

func (q *Queue) startConsumerLocked(queueName string, runtime *consumerRuntime) error {
	if err := q.ensureClientLocked(); err != nil {
		return err
	}
	stream := q.streamName(queueName)
	group := q.groupName(queueName)
	if err := q.createGroup(stream, group); err != nil {
		return err
	}
	runtime.started = true
	for i := 0; i < runtime.config.Concurrency; i++ {
		consumerName := runtime.consumerName + "-" + strconv.Itoa(i)
		go q.consumeLoop(queueName, stream, group, consumerName, runtime.config)
		go q.claimLoop(queueName, stream, group, consumerName, runtime.config)
	}
	return nil
}

func (q *Queue) createGroup(stream string, group string) error {
	err := q.rdb.XGroupCreateMkStream(context.Background(), stream, group, "0").Err()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (q *Queue) consumeLoop(queueName string, stream string, group string, consumer string, config asynctask.TaskQueueConfig) {
	for {
		select {
		case <-q.stopCh:
			return
		default:
		}
		res, err := q.rdb.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    int64(config.Concurrency),
			Block:    q.config.Block,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			time.Sleep(time.Second)
			continue
		}
		q.handleStreams(queueName, stream, group, res, config)
	}
}

func (q *Queue) claimLoop(queueName string, stream string, group string, consumer string, config asynctask.TaskQueueConfig) {
	ticker := time.NewTicker(q.config.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-q.stopCh:
			return
		case <-ticker.C:
			start := "0-0"
			for {
				result, next, err := q.rdb.XAutoClaim(context.Background(), &redis.XAutoClaimArgs{
					Stream:   stream,
					Group:    group,
					Consumer: consumer,
					MinIdle:  q.config.ClaimIdle,
					Start:    start,
					Count:    int64(config.Concurrency),
				}).Result()
				if err != nil {
					if errors.Is(err, redis.Nil) {
						break
					}
					break
				}
				if len(result) == 0 {
					break
				}
				q.handleMessages(queueName, stream, group, result, config)
				if next == start || next == "0-0" {
					break
				}
				start = next
			}
		}
	}
}

func (q *Queue) handleStreams(queueName string, stream string, group string, streams []redis.XStream, config asynctask.TaskQueueConfig) {
	for _, item := range streams {
		q.handleMessages(queueName, stream, group, item.Messages, config)
	}
}

func (q *Queue) handleMessages(queueName string, stream string, group string, messages []redis.XMessage, config asynctask.TaskQueueConfig) {
	for _, msg := range messages {
		task, err := decodeTask(msg)
		if err != nil {
			_, _ = q.rdb.XAck(context.Background(), stream, group, msg.ID).Result()
			continue
		}
		handler, err := q.lookupHandler(queueName, task.Type)
		if err != nil {
			_, _ = q.rdb.XAck(context.Background(), stream, group, msg.ID).Result()
			continue
		}
		if !task.AvailableAt.IsZero() && time.Now().Before(task.AvailableAt) {
			if err := q.requeueAfterDelay(task, time.Until(task.AvailableAt)); err == nil {
				_, _ = q.rdb.XAck(context.Background(), stream, group, msg.ID).Result()
			}
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
					err = fmt.Errorf("redisstream task panic: %v", recovered)
				}
			}()
			return handler.HandleTask(ctx, task)
		}()
		cancel()
		if handleErr == nil {
			task.Status = asynctask.QueueTaskStatusSucceeded
			_, _ = q.rdb.XAck(context.Background(), stream, group, msg.ID).Result()
			continue
		}

		task.LastError = handleErr.Error()
		maxRetry := config.MaxRetry
		if task.MaxRetry > 0 {
			maxRetry = task.MaxRetry
		}
		if task.Attempt >= maxRetry {
			task.Status = asynctask.QueueTaskStatusFailed
			_, _ = q.rdb.XAck(context.Background(), stream, group, msg.ID).Result()
			continue
		}

		task.Attempt++
		task.Status = asynctask.QueueTaskStatusRetrying
		if err := q.requeueAfterDelay(task, config.RetryDelay); err != nil {
			continue
		}
		_, _ = q.rdb.XAck(context.Background(), stream, group, msg.ID).Result()
	}
}

func (q *Queue) requeueAfterDelay(task *asynctask.QueueTask, delay time.Duration) error {
	if delay < 0 {
		delay = 0
	}
	task.Delay = delay
	task.AvailableAt = time.Now().Add(delay)
	_, err := q.Publish(context.Background(), task)
	return err
}

func decodeTask(msg redis.XMessage) (*asynctask.QueueTask, error) {
	raw, ok := msg.Values["task"]
	if !ok {
		return nil, asynctask.ErrInvalidQueuePayload.WithDetailStr("task payload missing")
	}
	payload, ok := raw.(string)
	if !ok {
		return nil, asynctask.ErrInvalidQueuePayload.WithDetailStr("task payload invalid")
	}
	task := &asynctask.QueueTask{}
	if err := json.Unmarshal([]byte(payload), task); err != nil {
		return nil, asynctask.ErrInvalidQueuePayload.WrapIfNot(err)
	}
	return task, nil
}

func (q *Queue) lookupHandler(queueName string, taskType string) (asynctask.TaskQueueHandler, error) {
	q.lock.RLock()
	defer q.lock.RUnlock()
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

func (q *Queue) streamName(queueName string) string {
	return q.config.StreamPrefix + queueName
}

func (q *Queue) groupName(queueName string) string {
	return q.config.GroupPrefix + queueName
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

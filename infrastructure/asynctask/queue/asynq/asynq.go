package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	libasynq "github.com/hibiken/asynq"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/datasource"
)

const (
	defaultShutdownTimeout = 30 * time.Second
)

type Config struct {
	Redis           datasource.RedisConfig
	ShutdownTimeout time.Duration
}

type handlerEntry struct {
	handler asynctask.TaskQueueHandler
}

type queueRuntime struct {
	config asynctask.TaskQueueConfig
	server *libasynq.Server
	mux    *libasynq.ServeMux
}

type Queue struct {
	config    Config
	lock      sync.Mutex
	reporter  asynctask.TaskReporter `aperture:"optional"`
	client    *libasynq.Client
	handlers  map[string]map[string]handlerEntry
	queues    map[string]*queueRuntime
	started   bool
	redisOpts libasynq.RedisClientOpt
}

func New(config Config) *Queue {
	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = defaultShutdownTimeout
	}
	opts := libasynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Username: config.Redis.Username,
		Password: config.Redis.Password,
		DB:       config.Redis.Database,
	}
	return &Queue{
		config:    config,
		handlers:  make(map[string]map[string]handlerEntry),
		queues:    make(map[string]*queueRuntime),
		redisOpts: opts,
	}
}

func (q *Queue) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskQueue", "hibiken/asynq")
}

func (q *Queue) Setup() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.client == nil {
		q.client = libasynq.NewClient(q.redisOpts)
	}
	for name, runtime := range q.queues {
		if runtime.server != nil {
			continue
		}
		if err := q.startQueueLocked(name, runtime); err != nil {
			return err
		}
	}
	q.started = true
	return nil
}

func (q *Queue) Dispose() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	for _, runtime := range q.queues {
		if runtime.server != nil {
			runtime.server.Shutdown()
		}
	}
	if q.client != nil {
		return q.client.Close()
	}
	return nil
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

	payload, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	internalTask := libasynq.NewTask(task.Type, payload)
	opts := []libasynq.Option{
		libasynq.Queue(task.Queue),
		libasynq.TaskID(task.ID),
	}
	if task.Delay > 0 {
		opts = append(opts, libasynq.ProcessIn(task.Delay))
	}
	if task.MaxRetry > 0 {
		opts = append(opts, libasynq.MaxRetry(task.MaxRetry))
	}
	if task.Timeout > 0 {
		opts = append(opts, libasynq.Timeout(task.Timeout))
	}

	q.lock.Lock()
	if q.client == nil {
		q.client = libasynq.NewClient(q.redisOpts)
	}
	client := q.client
	q.lock.Unlock()

	info, err := client.EnqueueContext(ctx, internalTask, opts...)
	if err != nil {
		return nil, err
	}
	if info != nil {
		task.ID = info.ID
	}
	q.report(ctx, q.event(task, asynctask.QueueTaskStatusPending, 0, ""))
	return task, nil
}

func (q *Queue) RegisterQueue(queueName string, config asynctask.TaskQueueConfig) error {
	if queueName == "" {
		return asynctask.ErrInvalidQueueName
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	normalized := normalizeConfig(config)
	runtime, ok := q.queues[queueName]
	if ok && !sameConfig(runtime.config, normalized) {
		return asynctask.ErrQueueConfigConflict.WithDetailStr(queueName)
	}
	if !ok {
		runtime = &queueRuntime{
			config: normalized,
			mux:    libasynq.NewServeMux(),
		}
		q.queues[queueName] = runtime
	}
	if q.started && runtime.server == nil {
		return q.startQueueLocked(queueName, runtime)
	}
	return nil
}

func (q *Queue) RegisterHandler(queueName string, taskType string, handler asynctask.TaskQueueHandler) error {
	if queueName == "" {
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

	runtime, ok := q.queues[queueName]
	if !ok {
		return asynctask.ErrQueueNotRegistered.WithDetailStr(queueName)
	}
	queueHandlers, ok := q.handlers[queueName]
	if !ok {
		queueHandlers = make(map[string]handlerEntry)
		q.handlers[queueName] = queueHandlers
	}
	queueHandlers[taskType] = handlerEntry{handler: handler}
	runtime.mux.HandleFunc(taskType, func(ctx context.Context, task *libasynq.Task) error {
		queueTask, err := decodeTask(task)
		if err != nil {
			return err
		}
		attempt, _ := libasynq.GetRetryCount(ctx)
		q.report(ctx, q.event(queueTask, asynctask.QueueTaskStatusRunning, attempt, ""))
		runErr := handler.HandleTask(ctx, queueTask)
		if runErr == nil {
			q.report(ctx, q.event(queueTask, asynctask.QueueTaskStatusSucceeded, attempt, ""))
			return nil
		}
		maxRetry, ok := libasynq.GetMaxRetry(ctx)
		if ok && attempt >= maxRetry {
			q.report(ctx, q.event(queueTask, asynctask.QueueTaskStatusFailed, attempt, runErr.Error()))
			return runErr
		}
		q.report(ctx, q.event(queueTask, asynctask.QueueTaskStatusRetrying, attempt+1, runErr.Error()))
		return runErr
	})
	return nil
}

func (q *Queue) startQueueLocked(queueName string, runtime *queueRuntime) error {
	server := libasynq.NewServer(q.redisOpts, libasynq.Config{
		Concurrency:     runtime.config.Concurrency,
		Queues:          map[string]int{queueName: 1},
		ShutdownTimeout: q.config.ShutdownTimeout,
	})
	if err := server.Start(runtime.mux); err != nil {
		return err
	}
	runtime.server = server
	return nil
}

func decodeTask(task *libasynq.Task) (*asynctask.QueueTask, error) {
	queueTask := &asynctask.QueueTask{}
	if err := json.Unmarshal(task.Payload(), queueTask); err != nil {
		return nil, asynctask.ErrInvalidQueuePayload.WrapIfNot(err)
	}
	if queueTask.ID == "" {
		queueTask.ID = uuid.NewString()
	}
	return queueTask, nil
}

func (q *Queue) event(task *asynctask.QueueTask, status asynctask.QueueTaskStatus, attempt int, err string) asynctask.TaskEvent {
	event := asynctask.NewTaskEvent(task)
	event.Status = status
	event.Attempt = attempt
	event.Error = err
	return event
}

func (q *Queue) report(ctx context.Context, event asynctask.TaskEvent) {
	_ = asynctask.ReportTaskEvent(ctx, q.reporter, event)
}

func normalizeConfig(config asynctask.TaskQueueConfig) asynctask.TaskQueueConfig {
	if config.Concurrency <= 0 {
		config.Concurrency = 1
	}
	return config
}

func sameConfig(left asynctask.TaskQueueConfig, right asynctask.TaskQueueConfig) bool {
	left = normalizeConfig(left)
	right = normalizeConfig(right)
	return left.Concurrency == right.Concurrency &&
		left.MaxRetry == right.MaxRetry &&
		left.RetryDelay == right.RetryDelay &&
		left.BufferSize == right.BufferSize
}

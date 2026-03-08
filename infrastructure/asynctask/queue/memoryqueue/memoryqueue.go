package memoryqueue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

const (
	defaultConcurrency = 1
	defaultRetryDelay  = time.Second
	defaultBufferSize  = 128
)

type handlerEntry struct {
	handler asynctask.TaskQueueHandler
}

type queueRuntime struct {
	config asynctask.TaskQueueConfig
	tasks  chan *asynctask.QueueTask
}

type MemoryQueue struct {
	lock     sync.RWMutex
	handlers map[string]map[string]handlerEntry
	queues   map[string]*queueRuntime
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		handlers: make(map[string]map[string]handlerEntry),
		queues:   make(map[string]*queueRuntime),
	}
}

func (m *MemoryQueue) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskQueue", "memory")
}

func (m *MemoryQueue) Publish(_ context.Context, task *asynctask.QueueTask) (*asynctask.QueueTask, error) {
	if task == nil || task.Queue == "" || task.Type == "" {
		return nil, asynctask.ErrInvalidQueueTask
	}
	task.Identifier()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.AvailableAt = task.CreatedAt.Add(task.Delay)
	task.Status = asynctask.QueueTaskStatusPending

	runtime := m.ensureQueue(task.Queue, asynctask.TaskQueueConfig{})
	if task.Delay > 0 {
		time.AfterFunc(task.Delay, func() {
			runtime.tasks <- task
		})
		return task, nil
	}
	runtime.tasks <- task
	return task, nil
}

func (m *MemoryQueue) RegisterQueue(queue string, config asynctask.TaskQueueConfig) error {
	if queue == "" {
		return asynctask.ErrInvalidQueueName
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	normalized := normalizeConfig(config)
	runtime, exists := m.queues[queue]
	if exists && !sameConfig(runtime.config, normalized) {
		return asynctask.ErrQueueConfigConflict.WithDetailStr(queue)
	}
	if !exists {
		m.startQueueLocked(queue, normalized)
	}
	return nil
}

func (m *MemoryQueue) RegisterHandler(queue string, taskType string, handler asynctask.TaskQueueHandler) error {
	if queue == "" {
		return asynctask.ErrInvalidQueueName
	}
	if taskType == "" {
		return asynctask.ErrInvalidTaskType
	}
	if handler == nil {
		return asynctask.ErrInvalidTaskHandler
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, exists := m.queues[queue]; !exists {
		return asynctask.ErrQueueNotRegistered.WithDetailStr(queue)
	}
	typeHandlers, ok := m.handlers[queue]
	if !ok {
		typeHandlers = make(map[string]handlerEntry)
		m.handlers[queue] = typeHandlers
	}
	typeHandlers[taskType] = handlerEntry{handler: handler}
	return nil
}

func (m *MemoryQueue) startQueueLocked(queue string, config asynctask.TaskQueueConfig) *queueRuntime {
	normalized := normalizeConfig(config)
	runtime := &queueRuntime{
		config: normalized,
		tasks:  make(chan *asynctask.QueueTask, normalized.BufferSize),
	}
	m.queues[queue] = runtime
	for i := 0; i < runtime.config.Concurrency; i++ {
		go m.worker(queue, runtime)
	}
	return runtime
}

func (m *MemoryQueue) ensureQueue(queue string, config asynctask.TaskQueueConfig) *queueRuntime {
	m.lock.Lock()
	defer m.lock.Unlock()
	if runtime, ok := m.queues[queue]; ok {
		return runtime
	}
	return m.startQueueLocked(queue, config)
}

func (m *MemoryQueue) worker(queue string, runtime *queueRuntime) {
	for task := range runtime.tasks {
		m.handle(queue, runtime, task)
	}
}

func (m *MemoryQueue) handle(queue string, runtime *queueRuntime, task *asynctask.QueueTask) {
	handler, err := m.lookupHandler(queue, task.Type)
	if err != nil {
		task.Status = asynctask.QueueTaskStatusFailed
		task.LastError = err.Error()
		return
	}

	task.Status = asynctask.QueueTaskStatusRunning
	ctx := context.Background()
	cancel := func() {}
	if task.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, task.Timeout)
	}
	err = handler.HandleTask(ctx, task)
	cancel()
	if err == nil {
		task.Status = asynctask.QueueTaskStatusSucceeded
		task.LastError = ""
		return
	}

	task.LastError = err.Error()
	maxRetry := runtime.config.MaxRetry
	if task.MaxRetry > 0 {
		maxRetry = task.MaxRetry
	}
	if task.Attempt >= maxRetry {
		task.Status = asynctask.QueueTaskStatusFailed
		return
	}

	task.Attempt++
	task.Status = asynctask.QueueTaskStatusRetrying
	delay := runtime.config.RetryDelay
	if delay <= 0 {
		delay = defaultRetryDelay
	}
	time.AfterFunc(delay, func() {
		task.Status = asynctask.QueueTaskStatusPending
		runtime.tasks <- task
	})
}

func (m *MemoryQueue) lookupHandler(queue string, taskType string) (asynctask.TaskQueueHandler, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	typeHandlers, ok := m.handlers[queue]
	if !ok {
		return nil, asynctask.ErrTaskHandlerNotFound.WithDetailStr(queue)
	}
	entry, ok := typeHandlers[taskType]
	if !ok {
		return nil, asynctask.ErrTaskHandlerNotFound.WithDetailStr(queue + ":" + taskType)
	}
	return entry.handler, nil
}

func normalizeConfig(config asynctask.TaskQueueConfig) asynctask.TaskQueueConfig {
	if config.Concurrency <= 0 {
		config.Concurrency = defaultConcurrency
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = defaultRetryDelay
	}
	if config.BufferSize <= 0 {
		config.BufferSize = defaultBufferSize
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

func IsHandlerNotFound(err error) bool {
	return errors.Is(err, asynctask.ErrTaskHandlerNotFound)
}

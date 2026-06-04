package memoryqueue

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
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
	config   asynctask.TaskQueueConfig
	lock     sync.Mutex
	capacity chan struct{}
	notify   chan struct{}
	ready    priorityTaskHeap
	delayed  delayedTaskHeap
	sequence int64
}

type MemoryQueue struct {
	lock     sync.RWMutex
	reporter asynctask.TaskReporter `aperture:"optional"`
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

func (m *MemoryQueue) Publish(ctx context.Context, task *asynctask.QueueTask) (*asynctask.QueueTask, error) {
	if task == nil || task.Queue == "" || task.Type == "" {
		return nil, asynctask.ErrInvalidQueueTask
	}
	task.Identifier()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.AvailableAt = task.CreatedAt.Add(task.Delay)

	runtime := m.ensureQueue(task.Queue, asynctask.TaskQueueConfig{})
	if err := runtime.push(ctx, task); err != nil {
		return nil, err
	}
	m.report(ctx, m.event(task, asynctask.QueueTaskStatusPending, 0, ""))
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
		config:   normalized,
		capacity: make(chan struct{}, normalized.BufferSize+normalized.Concurrency),
		notify:   make(chan struct{}, 1),
	}
	heap.Init(&runtime.ready)
	heap.Init(&runtime.delayed)
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
	for {
		item := runtime.pop()
		m.handle(queue, runtime, item)
	}
}

func (m *MemoryQueue) handle(queue string, runtime *queueRuntime, item priorityTask) {
	task := item.task
	handler, err := m.lookupHandler(queue, task.Type)
	if err != nil {
		m.report(context.Background(), m.event(task, asynctask.QueueTaskStatusFailed, item.attempt, err.Error()))
		runtime.release()
		return
	}

	m.report(context.Background(), m.event(task, asynctask.QueueTaskStatusRunning, item.attempt, ""))
	ctx := context.Background()
	cancel := func() {}
	if task.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, task.Timeout)
	}
	err = func() (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("memoryqueue task panic: %v", recovered)
			}
		}()
		return handler.HandleTask(ctx, task)
	}()
	cancel()
	if err == nil {
		m.report(context.Background(), m.event(task, asynctask.QueueTaskStatusSucceeded, item.attempt, ""))
		runtime.release()
		return
	}

	maxRetry := runtime.config.MaxRetry
	if task.MaxRetry > 0 {
		maxRetry = task.MaxRetry
	}
	if item.attempt >= maxRetry {
		m.report(context.Background(), m.event(task, asynctask.QueueTaskStatusFailed, item.attempt, err.Error()))
		runtime.release()
		return
	}

	nextAttempt := item.attempt + 1
	m.report(context.Background(), m.event(task, asynctask.QueueTaskStatusRetrying, nextAttempt, err.Error()))
	delay := runtime.config.RetryDelay
	if delay <= 0 {
		delay = defaultRetryDelay
	}
	task.Delay = delay
	task.AvailableAt = time.Now().Add(delay)
	runtime.requeue(task, nextAttempt)
}

func (r *queueRuntime) push(ctx context.Context, task *asynctask.QueueTask) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case r.capacity <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	r.enqueue(task, 0)
	return nil
}

func (r *queueRuntime) requeue(task *asynctask.QueueTask, attempt int) {
	r.enqueue(task, attempt)
}

func (r *queueRuntime) enqueue(task *asynctask.QueueTask, attempt int) {
	r.lock.Lock()
	r.sequence++
	item := priorityTask{task: task, attempt: attempt, sequence: r.sequence}
	if task.AvailableAt.IsZero() || !task.AvailableAt.After(time.Now()) {
		heap.Push(&r.ready, item)
	} else {
		heap.Push(&r.delayed, item)
	}
	r.lock.Unlock()
	r.signal()
}

func (r *queueRuntime) release() {
	select {
	case <-r.capacity:
	default:
	}
}

func (r *queueRuntime) pop() priorityTask {
	for {
		task, wait, hasDelayed := r.popReady()
		if task.task != nil {
			return task
		}
		if !hasDelayed {
			select {
			case <-r.notify:
			}
			continue
		}
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-r.notify:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}
	}
}

func (r *queueRuntime) popReady() (priorityTask, time.Duration, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	now := time.Now()
	for r.delayed.Len() > 0 {
		item := r.delayed[0]
		if item.task.AvailableAt.After(now) {
			break
		}
		heap.Push(&r.ready, heap.Pop(&r.delayed))
	}
	if r.ready.Len() > 0 {
		item := heap.Pop(&r.ready).(priorityTask)
		if r.ready.Len() > 0 {
			r.signal()
		}
		return item, 0, false
	}
	if r.delayed.Len() == 0 {
		return priorityTask{}, 0, false
	}
	wait := time.Until(r.delayed[0].task.AvailableAt)
	if wait <= 0 {
		return priorityTask{}, 0, true
	}
	return priorityTask{}, wait, true
}

func (r *queueRuntime) signal() {
	select {
	case r.notify <- struct{}{}:
	default:
	}
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

func (m *MemoryQueue) event(task *asynctask.QueueTask, status asynctask.QueueTaskStatus, attempt int, err string) asynctask.TaskEvent {
	event := asynctask.NewTaskEvent(task)
	event.Status = status
	event.Attempt = attempt
	event.Error = err
	return event
}

func (m *MemoryQueue) report(ctx context.Context, event asynctask.TaskEvent) {
	_ = asynctask.ReportTaskEvent(ctx, m.reporter, event)
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

type priorityTask struct {
	task     *asynctask.QueueTask
	attempt  int
	sequence int64
}

type priorityTaskHeap []priorityTask

func (h priorityTaskHeap) Len() int { return len(h) }

func (h priorityTaskHeap) Less(i, j int) bool {
	if h[i].task.Priority != h[j].task.Priority {
		return h[i].task.Priority > h[j].task.Priority
	}
	return h[i].sequence < h[j].sequence
}

func (h priorityTaskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *priorityTaskHeap) Push(x any) {
	*h = append(*h, x.(priorityTask))
}

func (h *priorityTaskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

type delayedTaskHeap []priorityTask

func (h delayedTaskHeap) Len() int { return len(h) }

func (h delayedTaskHeap) Less(i, j int) bool {
	left := h[i].task.AvailableAt
	right := h[j].task.AvailableAt
	if !left.Equal(right) {
		return left.Before(right)
	}
	if h[i].task.Priority != h[j].task.Priority {
		return h[i].task.Priority > h[j].task.Priority
	}
	return h[i].sequence < h[j].sequence
}

func (h delayedTaskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *delayedTaskHeap) Push(x any) {
	*h = append(*h, x.(priorityTask))
}

func (h *delayedTaskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

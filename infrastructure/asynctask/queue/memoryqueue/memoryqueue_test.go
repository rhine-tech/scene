package memoryqueue

import (
	"container/heap"
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

func newTestRuntime(bufferSize int) *queueRuntime {
	if bufferSize <= 0 {
		bufferSize = 1
	}
	runtime := &queueRuntime{
		config:   asynctask.TaskQueueConfig{Concurrency: 1, BufferSize: bufferSize},
		capacity: make(chan struct{}, bufferSize+1),
		notify:   make(chan struct{}, 1),
	}
	heap.Init(&runtime.ready)
	heap.Init(&runtime.delayed)
	return runtime
}

func TestMemoryQueuePublish(t *testing.T) {
	queue := NewMemoryQueue()
	done := make(chan struct{}, 1)

	err := queue.RegisterQueue("test.queue", asynctask.TaskQueueConfig{Concurrency: 1})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("test.queue", "echo", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		close(done)
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		Queue: "test.queue",
		Type:  "echo",
	}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("task not handled in time")
	}
}

func TestMemoryQueueRetry(t *testing.T) {
	queue := NewMemoryQueue()
	var attempts int32
	done := make(chan struct{}, 1)

	err := queue.RegisterQueue("retry.queue", asynctask.TaskQueueConfig{
		Concurrency: 1,
		MaxRetry:    2,
		RetryDelay:  10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("retry.queue", "retry", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		current := atomic.AddInt32(&attempts, 1)
		if current < 2 {
			return asynctask.ErrInternal
		}
		close(done)
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		Queue: "retry.queue",
		Type:  "retry",
	}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("task retry not completed in time")
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("unexpected attempt count: %d", got)
	}
}

func TestMemoryQueuePriority(t *testing.T) {
	queue := NewMemoryQueue()
	block := make(chan struct{})
	done := make(chan string, 3)

	err := queue.RegisterQueue("priority.queue", asynctask.TaskQueueConfig{Concurrency: 1})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("priority.queue", "record", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		if task.ID == "blocker" {
			<-block
		}
		done <- task.ID
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	for _, task := range []*asynctask.QueueTask{
		{ID: "blocker", Queue: "priority.queue", Type: "record", Priority: 100},
		{ID: "low", Queue: "priority.queue", Type: "record", Priority: 1},
		{ID: "high", Queue: "priority.queue", Type: "record", Priority: 9},
	} {
		if _, err := queue.Publish(context.Background(), task); err != nil {
			t.Fatalf("publish failed: %v", err)
		}
	}

	close(block)
	got := []string{}
	for len(got) < 3 {
		select {
		case id := <-done:
			got = append(got, id)
		case <-time.After(2 * time.Second):
			t.Fatalf("priority tasks not handled in time, got=%v", got)
		}
	}
	want := []string{"blocker", "high", "low"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected handling order: got=%v want=%v", got, want)
	}
}

func TestMemoryQueuePriorityFIFO(t *testing.T) {
	queue := NewMemoryQueue()
	block := make(chan struct{})
	done := make(chan string, 3)

	err := queue.RegisterQueue("fifo.queue", asynctask.TaskQueueConfig{Concurrency: 1})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("fifo.queue", "record", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		if task.ID == "blocker" {
			<-block
		}
		done <- task.ID
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	for _, task := range []*asynctask.QueueTask{
		{ID: "blocker", Queue: "fifo.queue", Type: "record", Priority: 100},
		{ID: "first", Queue: "fifo.queue", Type: "record", Priority: 5},
		{ID: "second", Queue: "fifo.queue", Type: "record", Priority: 5},
	} {
		if _, err := queue.Publish(context.Background(), task); err != nil {
			t.Fatalf("publish failed: %v", err)
		}
	}

	close(block)
	got := []string{}
	for len(got) < 3 {
		select {
		case id := <-done:
			got = append(got, id)
		case <-time.After(2 * time.Second):
			t.Fatalf("fifo tasks not handled in time, got=%v", got)
		}
	}
	want := []string{"blocker", "first", "second"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected handling order: got=%v want=%v", got, want)
	}
}

func TestMemoryQueueDelayedHighPriorityDoesNotBlockReadyTask(t *testing.T) {
	queue := NewMemoryQueue()
	done := make(chan string, 2)

	err := queue.RegisterQueue("delay-priority.queue", asynctask.TaskQueueConfig{Concurrency: 1})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("delay-priority.queue", "record", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		done <- task.ID
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:       "delayed-high",
		Queue:    "delay-priority.queue",
		Type:     "record",
		Priority: 100,
		Delay:    50 * time.Millisecond,
	}); err != nil {
		t.Fatalf("publish delayed task failed: %v", err)
	}
	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:       "ready-low",
		Queue:    "delay-priority.queue",
		Type:     "record",
		Priority: 1,
	}); err != nil {
		t.Fatalf("publish ready task failed: %v", err)
	}

	select {
	case got := <-done:
		if got != "ready-low" {
			t.Fatalf("unexpected first task: got=%s want=ready-low", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ready task not handled in time")
	}
	select {
	case got := <-done:
		if got != "delayed-high" {
			t.Fatalf("unexpected second task: got=%s want=delayed-high", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("delayed task not handled in time")
	}
}

func TestQueueRuntimePopReadySignalsWhenReadyTasksRemain(t *testing.T) {
	runtime := newTestRuntime(2)
	now := time.Now()
	for _, id := range []string{"first", "second"} {
		if err := runtime.push(context.Background(), &asynctask.QueueTask{
			ID:          id,
			Queue:       "signal.queue",
			Type:        "record",
			AvailableAt: now,
		}); err != nil {
			t.Fatalf("push failed: %v", err)
		}
	}

	for {
		select {
		case <-runtime.notify:
		default:
			goto drained
		}
	}

drained:
	task, _, _ := runtime.popReady()
	if task == nil {
		t.Fatal("expected ready task")
	}
	select {
	case <-runtime.notify:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected remaining ready task notification")
	}
}

func TestMemoryQueueRetryDoesNotBlockWorkerWhenQueueIsFull(t *testing.T) {
	queue := NewMemoryQueue()
	var attempts int32
	done := make(chan struct{})
	retryStarted := make(chan struct{})
	proceedRetry := make(chan struct{})

	err := queue.RegisterQueue("retry-full.queue", asynctask.TaskQueueConfig{
		Concurrency: 1,
		BufferSize:  1,
		MaxRetry:    1,
		RetryDelay:  10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("retry-full.queue", "retry", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		current := atomic.AddInt32(&attempts, 1)
		if current == 1 {
			close(retryStarted)
			<-proceedRetry
			return asynctask.ErrInternal
		}
		close(done)
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:    "retrying",
		Queue: "retry-full.queue",
		Type:  "retry",
	}); err != nil {
		t.Fatalf("publish retry task failed: %v", err)
	}
	select {
	case <-retryStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("retry task did not start")
	}
	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:    "filler",
		Queue: "retry-full.queue",
		Type:  "missing",
		Delay: time.Hour,
	}); err != nil {
		t.Fatalf("publish filler task failed: %v", err)
	}
	close(proceedRetry)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("retry was blocked by full queue")
	}
}

func TestMemoryQueuePublishReturnsContextErrorWhenQueueIsFull(t *testing.T) {
	queue := NewMemoryQueue()
	block := make(chan struct{})

	err := queue.RegisterQueue("full.queue", asynctask.TaskQueueConfig{
		Concurrency: 1,
		BufferSize:  1,
	})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("full.queue", "block", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		<-block
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:    "active",
		Queue: "full.queue",
		Type:  "block",
	}); err != nil {
		t.Fatalf("publish active task failed: %v", err)
	}
	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:    "queued",
		Queue: "full.queue",
		Type:  "block",
	}); err != nil {
		t.Fatalf("publish queued task failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := queue.Publish(ctx, &asynctask.QueueTask{
		ID:    "overflow",
		Queue: "full.queue",
		Type:  "block",
	}); err == nil {
		t.Fatal("expected context error when queue is full")
	}
	close(block)
}

func TestMemoryQueueHandlerPanicDoesNotStopWorker(t *testing.T) {
	queue := NewMemoryQueue()
	done := make(chan struct{})

	err := queue.RegisterQueue("panic.queue", asynctask.TaskQueueConfig{
		Concurrency: 1,
		BufferSize:  1,
	})
	if err != nil {
		t.Fatalf("register queue failed: %v", err)
	}
	err = queue.RegisterHandler("panic.queue", "handle", asynctask.TaskQueueHandlerFunc(func(ctx context.Context, task *asynctask.QueueTask) error {
		if task.ID == "panic" {
			panic("boom")
		}
		close(done)
		return nil
	}))
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:    "panic",
		Queue: "panic.queue",
		Type:  "handle",
	}); err != nil {
		t.Fatalf("publish panic task failed: %v", err)
	}
	if _, err := queue.Publish(context.Background(), &asynctask.QueueTask{
		ID:    "after-panic",
		Queue: "panic.queue",
		Type:  "handle",
	}); err != nil {
		t.Fatalf("publish follow-up task failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker stopped after handler panic")
	}
}

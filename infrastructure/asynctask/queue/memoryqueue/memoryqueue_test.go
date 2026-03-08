package memoryqueue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

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

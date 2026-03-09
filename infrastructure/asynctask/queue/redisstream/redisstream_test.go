package redisstream

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/datasource"
)

func newTestQueue(t *testing.T) *Queue {
	t.Helper()
	queue := New(Config{
		Redis: datasource.DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     6379,
			Database: "15",
		},
		StreamPrefix: "scene.test.asynctask.stream.",
		GroupPrefix:  "scene.test.asynctask.group.",
		Block:        100 * time.Millisecond,
		ClaimIdle:    200 * time.Millisecond,
		PollInterval: 100 * time.Millisecond,
	})
	if err := queue.Setup(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() {
		if queue.rdb != nil {
			_ = queue.rdb.FlushDB(context.Background()).Err()
		}
		_ = queue.Dispose()
	})
	return queue
}

func TestRedisStreamPublishConsume(t *testing.T) {
	queue := newTestQueue(t)
	done := make(chan struct{}, 1)

	err := queue.RegisterQueue("test.queue", asynctask.TaskQueueConfig{
		Concurrency: 1,
		MaxRetry:    1,
		RetryDelay:  10 * time.Millisecond,
	})
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
	case <-time.After(3 * time.Second):
		t.Fatal("task not handled in time")
	}
}

func TestRedisStreamRetry(t *testing.T) {
	queue := newTestQueue(t)
	done := make(chan struct{}, 1)
	var attempts int32

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
	case <-time.After(3 * time.Second):
		t.Fatal("retry task not handled in time")
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("unexpected attempt count: %d", got)
	}
}

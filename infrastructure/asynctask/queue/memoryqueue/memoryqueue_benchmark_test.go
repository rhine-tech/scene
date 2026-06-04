package memoryqueue

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

func BenchmarkQueueRuntimePushReady(b *testing.B) {
	runtime := newTestRuntime(b.N)
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runtime.push(context.Background(), &asynctask.QueueTask{
			Queue:       "bench.queue",
			Type:        "bench",
			Priority:    i % 10,
			AvailableAt: now,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueueRuntimePushDelayed(b *testing.B) {
	runtime := newTestRuntime(b.N)
	availableAt := time.Now().Add(time.Hour)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runtime.push(context.Background(), &asynctask.QueueTask{
			Queue:       "bench.queue",
			Type:        "bench",
			Priority:    i % 10,
			AvailableAt: availableAt.Add(time.Duration(i) * time.Millisecond),
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQueueRuntimePopReady(b *testing.B) {
	runtime := newTestRuntime(b.N)
	now := time.Now()
	for i := 0; i < b.N; i++ {
		err := runtime.push(context.Background(), &asynctask.QueueTask{
			Queue:       "bench.queue",
			Type:        "bench",
			Priority:    i % 10,
			AvailableAt: now,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, _, _ := runtime.popReady()
		if task.task == nil {
			b.Fatal("expected ready task")
		}
	}
}

package memoryreporter

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/asynctask"
	v2reporter "github.com/rhine-tech/scene/infrastructure/asynctask/reporter/memoryreporter/v2"
)

func TestMemoryReporterFacadeUsesV2(t *testing.T) {
	reporter := NewMemoryReporter()
	v2Reporter := v2reporter.NewMemoryReporter()
	if reporter.ImplName() != v2Reporter.ImplName() {
		t.Fatalf("facade should use v2 implementation, got %s want %s", reporter.ImplName(), v2Reporter.ImplName())
	}
}

func TestMemoryReporterFacadeReportAndGetTask(t *testing.T) {
	reporter := NewMemoryReporter()
	now := time.Now()
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
		TaskID:    "task-1",
		Queue:     "queue",
		Status:    asynctask.QueueTaskStatusRunning,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("report task event failed: %v", err)
	}

	record, err := reporter.GetTask(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if record.TaskID != "task-1" || record.Queue != "queue" || record.Status != asynctask.QueueTaskStatusRunning {
		t.Fatalf("unexpected record: %+v", record)
	}
}

func BenchmarkMemoryReporterConcurrentReport(b *testing.B) {
	reporter := NewMemoryReporter()
	taskIDs := make([]string, 1024)
	for i := range taskIDs {
		taskIDs[i] = "task-" + strconv.Itoa(i)
	}

	var sequence atomic.Int64
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := sequence.Add(1)
			taskID := taskIDs[int(i)%len(taskIDs)]
			if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
				TaskID:    taskID,
				Queue:     "queue",
				TaskType:  "type",
				Status:    asynctask.QueueTaskStatusRunning,
				Attempt:   int(i),
				Progress:  int(i % 100),
				UpdatedAt: time.Unix(0, i),
			}); err != nil {
				b.Fatalf("report task event failed: %v", err)
			}
		}
	})
}

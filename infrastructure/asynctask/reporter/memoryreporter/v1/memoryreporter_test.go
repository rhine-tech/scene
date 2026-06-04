package v1

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

func TestMemoryReporterReportGetListStats(t *testing.T) {
	reporter := NewMemoryReporter()
	now := time.Now()
	for _, event := range []asynctask.TaskEvent{
		{TaskID: "task-1", Queue: "queue-a", TaskType: "type-a", Key: "key-a", Status: asynctask.QueueTaskStatusPending, UpdatedAt: now.Add(-2 * time.Second)},
		{TaskID: "task-1", Status: asynctask.QueueTaskStatusRunning, Attempt: 2, Progress: 50, UpdatedAt: now.Add(-time.Second)},
		{TaskID: "task-2", Queue: "queue-a", TaskType: "type-b", Status: asynctask.QueueTaskStatusFailed, Error: "failed", UpdatedAt: now},
	} {
		if err := reporter.ReportTaskEvent(context.Background(), event); err != nil {
			t.Fatalf("report event failed: %v", err)
		}
	}

	record, err := reporter.GetTask(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if record.Queue != "queue-a" || record.TaskType != "type-a" || record.Key != "key-a" {
		t.Fatalf("unexpected identity: %+v", record)
	}
	if record.Status != asynctask.QueueTaskStatusRunning || record.Attempt != 2 || record.Progress != 50 {
		t.Fatalf("unexpected task state: %+v", record)
	}

	list, err := reporter.ListTasks(context.Background(), asynctask.TaskQuery{
		Queue:  "queue-a",
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("list tasks failed: %v", err)
	}
	if list.Total != 2 || list.Results[0].TaskID != "task-2" || list.Results[1].TaskID != "task-1" {
		t.Fatalf("unexpected list result: %+v", list)
	}

	stats, err := reporter.GetQueueStats(context.Background(), "queue-a")
	if err != nil {
		t.Fatalf("get queue stats failed: %v", err)
	}
	if stats.Total != 2 || stats.ByStatus[asynctask.QueueTaskStatusRunning] != 1 || stats.ByStatus[asynctask.QueueTaskStatusFailed] != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestMemoryReporterRetention(t *testing.T) {
	reporter := NewMemoryReporter(Config{EventTTL: time.Second, MaxTasks: -1, CleanupInterval: -1})
	now := time.Now()
	for _, event := range []asynctask.TaskEvent{
		{TaskID: "old", Queue: "queue", Status: asynctask.QueueTaskStatusSucceeded, UpdatedAt: now.Add(-2 * time.Second)},
		{TaskID: "new", Queue: "queue", Status: asynctask.QueueTaskStatusRunning, UpdatedAt: now},
	} {
		if err := reporter.ReportTaskEvent(context.Background(), event); err != nil {
			t.Fatalf("report event failed: %v", err)
		}
	}

	if removed := reporter.Cleanup(context.Background()); removed != 1 {
		t.Fatalf("unexpected removed count: %d", removed)
	}
	if _, err := reporter.GetTask(context.Background(), "old"); !errors.Is(err, asynctask.ErrTaskNotFound) {
		t.Fatalf("old task should be expired: %v", err)
	}
	if _, err := reporter.GetTask(context.Background(), "new"); err != nil {
		t.Fatalf("new task should be retained: %v", err)
	}
}

func TestMemoryReporterMaxTasksEvictsOldest(t *testing.T) {
	reporter := NewMemoryReporter(Config{EventTTL: -1, MaxTasks: 2, CleanupInterval: -1})
	now := time.Now()
	for i, taskID := range []string{"task-a", "task-b", "task-c"} {
		if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
			TaskID:    taskID,
			Queue:     "queue",
			Status:    asynctask.QueueTaskStatusPending,
			UpdatedAt: now.Add(time.Duration(i) * time.Second),
		}); err != nil {
			t.Fatalf("report event failed: %v", err)
		}
	}

	if _, err := reporter.GetTask(context.Background(), "task-a"); !errors.Is(err, asynctask.ErrTaskNotFound) {
		t.Fatalf("oldest task should be evicted: %v", err)
	}
}

func TestMemoryReporterLargeOffsetReturnsEmptyPage(t *testing.T) {
	reporter := NewMemoryReporter()
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
		TaskID:    "task-1",
		Queue:     "queue",
		Status:    asynctask.QueueTaskStatusPending,
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("report event failed: %v", err)
	}

	result, err := reporter.ListTasks(context.Background(), asynctask.TaskQuery{
		Offset: 1<<62 - 1,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("list tasks failed: %v", err)
	}
	if result.Offset != 1 || result.Count != 0 {
		t.Fatalf("unexpected large offset result: %+v", result)
	}
}

func TestMemoryReporterOutOfOrderAndPartialEvent(t *testing.T) {
	reporter := NewMemoryReporter()
	now := time.Now()
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
		TaskID:    "task-1",
		Status:    asynctask.QueueTaskStatusSucceeded,
		Attempt:   3,
		Message:   "done",
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("report latest event failed: %v", err)
	}
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
		TaskID:    "task-1",
		Status:    asynctask.QueueTaskStatusRunning,
		Message:   "stale",
		UpdatedAt: now.Add(-time.Second),
	}); err != nil {
		t.Fatalf("report stale event failed: %v", err)
	}
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
		TaskID:    "task-1",
		Progress:  80,
		UpdatedAt: now.Add(time.Second),
	}); err != nil {
		t.Fatalf("report progress event failed: %v", err)
	}

	record, err := reporter.GetTask(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if record.Status != asynctask.QueueTaskStatusSucceeded || record.Attempt != 3 || record.Progress != 80 || record.Message != "done" {
		t.Fatalf("unexpected merged record: %+v", record)
	}
}

func TestMemoryReporterErrors(t *testing.T) {
	reporter := NewMemoryReporter()
	if err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{}); !errors.Is(err, asynctask.ErrInvalidQueueTask) {
		t.Fatalf("unexpected report error: %v", err)
	}
	if _, err := reporter.GetTask(context.Background(), "missing"); !errors.Is(err, asynctask.ErrTaskNotFound) {
		t.Fatalf("unexpected get error: %v", err)
	}
}

func TestMemoryReporterConcurrentReport(t *testing.T) {
	reporter := NewMemoryReporter()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := reporter.ReportTaskEvent(context.Background(), asynctask.TaskEvent{
				TaskID:    "task-" + strconv.Itoa(i%10),
				Queue:     "queue",
				TaskType:  "type",
				Status:    asynctask.QueueTaskStatusRunning,
				Attempt:   i,
				Progress:  i % 100,
				UpdatedAt: time.Now().Add(time.Duration(i) * time.Millisecond),
			})
			if err != nil {
				t.Errorf("report task event failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	result, err := reporter.ListTasks(context.Background(), asynctask.TaskQuery{Queue: "queue"})
	if err != nil {
		t.Fatalf("list tasks failed: %v", err)
	}
	if result.Total != 10 {
		t.Fatalf("unexpected task count: %d", result.Total)
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

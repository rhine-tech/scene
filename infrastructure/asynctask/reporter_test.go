package asynctask

import (
	"context"
	"testing"

	"github.com/rhine-tech/scene"
)

type fakeTaskReporter struct {
	events []TaskEvent
}

func (f *fakeTaskReporter) ImplName() scene.ImplName {
	return Lens.ImplName("TaskReporter", "fake")
}

func (f *fakeTaskReporter) ReportTaskEvent(ctx context.Context, event TaskEvent) error {
	f.events = append(f.events, event)
	return nil
}

func TestReportTaskProgress(t *testing.T) {
	reporter := &fakeTaskReporter{}
	task := &QueueTask{
		ID:    "task-1",
		Queue: "queue",
		Type:  "type",
		Key:   "key",
	}

	err := ReportTaskProgress(context.Background(), reporter, task, 40, "processing")
	if err != nil {
		t.Fatalf("report task progress failed: %v", err)
	}
	if len(reporter.events) != 1 {
		t.Fatalf("unexpected event count: %d", len(reporter.events))
	}
	event := reporter.events[0]
	if event.TaskID != task.ID || event.Queue != task.Queue || event.TaskType != task.Type || event.Key != task.Key {
		t.Fatalf("unexpected task identity in event: %+v", event)
	}
	if event.Status != QueueTaskStatusRunning {
		t.Fatalf("unexpected status: %s", event.Status)
	}
	if event.Progress != 40 || event.Message != "processing" {
		t.Fatalf("unexpected progress event: %+v", event)
	}
	if event.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be set")
	}
}

func TestReportTaskEventIgnoresNilReporter(t *testing.T) {
	err := ReportTaskEvent(context.Background(), nil, TaskEvent{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

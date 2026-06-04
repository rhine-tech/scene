package asynctask

import (
	"context"
	"time"

	"github.com/rhine-tech/scene"
)

type TaskEvent struct {
	TaskID    string          `json:"task_id"`
	Queue     string          `json:"queue"`
	TaskType  string          `json:"task_type"`
	Key       string          `json:"key,omitempty"`
	Status    QueueTaskStatus `json:"status"`
	Attempt   int             `json:"attempt,omitempty"`
	Progress  int             `json:"progress,omitempty"`
	Message   string          `json:"message,omitempty"`
	Error     string          `json:"error,omitempty"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type TaskReporter interface {
	scene.Named
	ReportTaskEvent(ctx context.Context, event TaskEvent) error
}

func NewTaskEvent(task *QueueTask) TaskEvent {
	event := TaskEvent{UpdatedAt: time.Now()}
	if task == nil {
		return event
	}
	event.TaskID = task.ID
	event.Queue = task.Queue
	event.TaskType = task.Type
	event.Key = task.Key
	return event
}

func ReportTaskEvent(ctx context.Context, reporter TaskReporter, event TaskEvent) error {
	if reporter == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = time.Now()
	}
	return reporter.ReportTaskEvent(ctx, event)
}

func ReportTaskProgress(ctx context.Context, reporter TaskReporter, task *QueueTask, progress int, message string) error {
	event := NewTaskEvent(task)
	event.Status = QueueTaskStatusRunning
	event.Progress = progress
	event.Message = message
	return ReportTaskEvent(ctx, reporter, event)
}

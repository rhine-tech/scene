package asynctask

import (
	"context"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
)

type TaskQuery struct {
	Offset      int64             `json:"offset,omitempty"`
	Limit       int64             `json:"limit,omitempty"`
	TaskID      string            `json:"task_id,omitempty"`
	Queue       string            `json:"queue,omitempty"`
	TaskType    string            `json:"task_type,omitempty"`
	Key         string            `json:"key,omitempty"`
	Status      []QueueTaskStatus `json:"status,omitempty"`
	Keyword     string            `json:"keyword,omitempty"`
	UpdatedFrom time.Time         `json:"updated_from,omitempty"`
	UpdatedTo   time.Time         `json:"updated_to,omitempty"`
}

type QueueStats struct {
	Queue    string                    `json:"queue"`
	Total    int64                     `json:"total"`
	ByStatus map[QueueTaskStatus]int64 `json:"by_status"`
}

type TaskQueueInspector interface {
	scene.Named
	ListTasks(ctx context.Context, query TaskQuery) (*model.PaginationResult[TaskEvent], error)
	GetTask(ctx context.Context, taskID string) (TaskEvent, error)
	GetQueueStats(ctx context.Context, queue string) (QueueStats, error)
}

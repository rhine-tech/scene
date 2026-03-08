package asynctask

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
)

type QueueTaskStatus string

const (
	QueueTaskStatusPending   QueueTaskStatus = "pending"
	QueueTaskStatusRunning   QueueTaskStatus = "running"
	QueueTaskStatusRetrying  QueueTaskStatus = "retrying"
	QueueTaskStatusSucceeded QueueTaskStatus = "succeeded"
	QueueTaskStatusFailed    QueueTaskStatus = "failed"
)

type QueueTask struct {
	ID string `json:"id"`
	// Queue is the logical queue name.
	// It defines the consumption domain and queue-level runtime config.
	Queue string `json:"queue"`
	// Type is the task type inside a queue.
	// Consumers use it to route the task to a specific handler.
	Type string `json:"type"`
	// Key is an optional business grouping key.
	// It can be used by backends or future schedulers for partitioning,
	// deduplication, or serializing tasks for the same business object.
	Key string `json:"key,omitempty"`
	// Payload is the serialized task body, usually JSON.
	Payload []byte `json:"payload,omitempty"`
	// Headers stores optional metadata for tracing or backend-specific routing.
	Headers map[string]string `json:"headers,omitempty"`
	// Delay requests delayed execution.
	// The exact behavior depends on the backend implementation.
	Delay time.Duration `json:"delay,omitempty"`
	// Timeout is the handler execution timeout for this task.
	Timeout time.Duration `json:"timeout,omitempty"`
	// MaxRetry overrides the queue-level retry limit when greater than zero.
	MaxRetry int `json:"max_retry,omitempty"`
	// Attempt is the current retry attempt count.
	Attempt     int             `json:"attempt,omitempty"`
	Status      QueueTaskStatus `json:"status,omitempty"`
	LastError   string          `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	AvailableAt time.Time       `json:"available_at"`
}

func (t *QueueTask) Identifier() string {
	if t.ID != "" {
		return t.ID
	}
	t.ID = uuid.New().String()
	return t.ID
}

type TaskQueueConfig struct {
	// Concurrency is the number of workers consuming the logical queue.
	Concurrency int
	// MaxRetry is the default retry limit for tasks in the queue.
	MaxRetry int
	// RetryDelay is the default delay before retrying a failed task.
	RetryDelay time.Duration
	// BufferSize is the in-memory queue buffer size.
	// It is only meaningful for in-process backends such as memoryqueue.
	BufferSize int
}

// TaskQueuePublisher publishes queue tasks into a backend implementation.
// Implementations should treat Queue and Type as required routing metadata.
type TaskQueuePublisher interface {
	scene.Named
	// Publish enqueues a task to the backend and returns the stored task model.
	// The implementation may populate fields such as ID, CreatedAt and Status.
	Publish(ctx context.Context, task *QueueTask) (*QueueTask, error)
}

// TaskQueueHandler processes a single queue task delivered by a consumer.
type TaskQueueHandler interface {
	HandleTask(ctx context.Context, task *QueueTask) error
}

type TaskQueueHandlerFunc func(ctx context.Context, task *QueueTask) error

func (f TaskQueueHandlerFunc) HandleTask(ctx context.Context, task *QueueTask) error {
	return f(ctx, task)
}

// TaskQueueConsumer registers queues and task handlers, then starts consuming from queues.
type TaskQueueConsumer interface {
	scene.Named
	// RegisterQueue defines the config of a logical queue.
	// Re-registering the same queue requires an identical config.
	RegisterQueue(queue string, config TaskQueueConfig) error
	// RegisterHandler binds a task type to a logical queue.
	// The queue must be registered before handlers are attached.
	RegisterHandler(queue string, taskType string, handler TaskQueueHandler) error
}

func MarshalPayload(payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}
	return json.Marshal(payload)
}

func UnmarshalPayload(task *QueueTask, payload any) error {
	if task == nil || len(task.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(task.Payload, payload)
}

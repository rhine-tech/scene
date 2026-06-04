package asynctask

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/rhine-tech/scene"
)

type fakeQueueConsumer struct {
	queueErr   error
	handlerErr error
	calls      []string
	queues     []registeredQueue
	handlers   []registeredHandler
}

type registeredQueue struct {
	queue  string
	config TaskQueueConfig
}

type registeredHandler struct {
	queue    string
	taskType string
	handler  TaskQueueHandler
}

func (f *fakeQueueConsumer) ImplName() scene.ImplName {
	return Lens.ImplName("TaskQueue", "fake")
}

func (f *fakeQueueConsumer) RegisterQueue(queue string, config TaskQueueConfig) error {
	f.calls = append(f.calls, "queue:"+queue)
	if f.queueErr != nil {
		return f.queueErr
	}
	f.queues = append(f.queues, registeredQueue{queue: queue, config: config})
	return nil
}

func (f *fakeQueueConsumer) RegisterHandler(queue string, taskType string, handler TaskQueueHandler) error {
	f.calls = append(f.calls, "handler:"+queue+":"+taskType)
	if f.handlerErr != nil {
		return f.handlerErr
	}
	f.handlers = append(f.handlers, registeredHandler{
		queue:    queue,
		taskType: taskType,
		handler:  handler,
	})
	return nil
}

func TestRegisterQueueSpecs(t *testing.T) {
	consumer := &fakeQueueConsumer{}
	handler := TaskQueueHandlerFunc(func(ctx context.Context, task *QueueTask) error {
		return nil
	})
	config := TaskQueueConfig{
		Concurrency: 2,
		MaxRetry:    3,
		RetryDelay:  10 * time.Millisecond,
		BufferSize:  8,
	}

	err := RegisterQueueSpecs(consumer,
		QueueSpec{
			Queue:  " first.queue ",
			Config: config,
			Handlers: []HandlerSpec{
				{Type: " first.task ", Handler: handler},
			},
		},
		QueueSpec{
			Queue: "second.queue",
			Handlers: []HandlerSpec{
				{Type: "second.task", Handler: handler},
			},
		},
	)
	if err != nil {
		t.Fatalf("register queue specs failed: %v", err)
	}

	wantCalls := []string{
		"queue:first.queue",
		"handler:first.queue:first.task",
		"queue:second.queue",
		"handler:second.queue:second.task",
	}
	if !reflect.DeepEqual(consumer.calls, wantCalls) {
		t.Fatalf("unexpected calls: got=%v want=%v", consumer.calls, wantCalls)
	}
	if len(consumer.queues) != 2 {
		t.Fatalf("unexpected queue count: %d", len(consumer.queues))
	}
	if consumer.queues[0].config != config {
		t.Fatalf("unexpected queue config: got=%+v want=%+v", consumer.queues[0].config, config)
	}
	if len(consumer.handlers) != 2 {
		t.Fatalf("unexpected handler count: %d", len(consumer.handlers))
	}
}

func TestRegisterQueueSpecsValidation(t *testing.T) {
	handler := TaskQueueHandlerFunc(func(ctx context.Context, task *QueueTask) error {
		return nil
	})
	tests := []struct {
		name     string
		consumer TaskQueueConsumer
		specs    []QueueSpec
		wantErr  error
	}{
		{
			name:    "nil consumer",
			wantErr: ErrInvalidTaskQueueConsumer,
		},
		{
			name:     "empty queue",
			consumer: &fakeQueueConsumer{},
			specs: []QueueSpec{
				{Queue: " ", Handlers: []HandlerSpec{{Type: "task", Handler: handler}}},
			},
			wantErr: ErrInvalidQueueName,
		},
		{
			name:     "empty task type",
			consumer: &fakeQueueConsumer{},
			specs: []QueueSpec{
				{Queue: "queue", Handlers: []HandlerSpec{{Type: " ", Handler: handler}}},
			},
			wantErr: ErrInvalidTaskType,
		},
		{
			name:     "nil handler",
			consumer: &fakeQueueConsumer{},
			specs: []QueueSpec{
				{Queue: "queue", Handlers: []HandlerSpec{{Type: "task"}}},
			},
			wantErr: ErrInvalidTaskHandler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterQueueSpecs(tt.consumer, tt.specs...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("unexpected error: got=%v want=%v", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterQueueSpecsReturnsConsumerError(t *testing.T) {
	queueErr := errors.New("register queue failed")
	consumer := &fakeQueueConsumer{queueErr: queueErr}

	err := RegisterQueueSpecs(consumer, QueueSpec{Queue: "queue"})
	if !errors.Is(err, queueErr) {
		t.Fatalf("unexpected error: got=%v want=%v", err, queueErr)
	}

	handlerErr := errors.New("register handler failed")
	consumer = &fakeQueueConsumer{handlerErr: handlerErr}
	err = RegisterQueueSpecs(consumer, QueueSpec{
		Queue: "queue",
		Handlers: []HandlerSpec{
			{
				Type: "task",
				Handler: TaskQueueHandlerFunc(func(ctx context.Context, task *QueueTask) error {
					return nil
				}),
			},
		},
	})
	if !errors.Is(err, handlerErr) {
		t.Fatalf("unexpected error: got=%v want=%v", err, handlerErr)
	}
}

func TestPayloadHandler(t *testing.T) {
	type payload struct {
		ID string `json:"id"`
	}
	body, err := MarshalPayload(payload{ID: "p1"})
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	var gotPayload payload
	var gotTask *QueueTask
	handler := PayloadHandler(func(ctx context.Context, task *QueueTask, p payload) error {
		gotTask = task
		gotPayload = p
		return nil
	})
	task := &QueueTask{Payload: body}
	if err := handler.HandleTask(context.Background(), task); err != nil {
		t.Fatalf("handle task failed: %v", err)
	}
	if gotTask != task {
		t.Fatal("handler did not receive original task")
	}
	if gotPayload.ID != "p1" {
		t.Fatalf("unexpected payload: %+v", gotPayload)
	}
}

func TestPayloadHandlerReturnsUnmarshalError(t *testing.T) {
	handler := PayloadHandler(func(ctx context.Context, task *QueueTask, payload struct{}) error {
		return nil
	})
	err := handler.HandleTask(context.Background(), &QueueTask{Payload: []byte("{")})
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

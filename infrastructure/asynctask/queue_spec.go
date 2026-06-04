package asynctask

import (
	"context"
	"strings"
)

// QueueSpec describes one logical task queue and the handlers owned by a worker.
type QueueSpec struct {
	// Queue is the logical queue name registered on TaskQueueConsumer.
	Queue string
	// Config is the queue-level runtime configuration.
	Config TaskQueueConfig
	// Handlers binds task types to handlers in this queue.
	Handlers []HandlerSpec
}

// HandlerSpec describes one task type handler inside a QueueSpec.
type HandlerSpec struct {
	// Type is the task type routed by TaskQueueConsumer.
	Type string
	// Handler processes tasks of Type.
	Handler TaskQueueHandler
}

// RegisterQueueSpecs registers queues first, then all handlers in each queue.
// It is intended to be called from a module-owned worker app, usually in Run().
func RegisterQueueSpecs(consumer TaskQueueConsumer, specs ...QueueSpec) error {
	if consumer == nil {
		return ErrInvalidTaskQueueConsumer
	}
	for _, spec := range specs {
		queueName := strings.TrimSpace(spec.Queue)
		if queueName == "" {
			return ErrInvalidQueueName
		}
		if err := consumer.RegisterQueue(queueName, spec.Config); err != nil {
			return err
		}
		for _, handlerSpec := range spec.Handlers {
			taskType := strings.TrimSpace(handlerSpec.Type)
			if taskType == "" {
				return ErrInvalidTaskType
			}
			if handlerSpec.Handler == nil {
				return ErrInvalidTaskHandler
			}
			if err := consumer.RegisterHandler(queueName, taskType, handlerSpec.Handler); err != nil {
				return err
			}
		}
	}
	return nil
}

// PayloadHandler unmarshals QueueTask.Payload into T before invoking handler.
func PayloadHandler[T any](handler func(ctx context.Context, task *QueueTask, payload T) error) TaskQueueHandler {
	return TaskQueueHandlerFunc(func(ctx context.Context, task *QueueTask) error {
		if handler == nil {
			return ErrInvalidTaskHandler
		}
		var payload T
		if err := UnmarshalPayload(task, &payload); err != nil {
			return err
		}
		return handler(ctx, task, payload)
	})
}

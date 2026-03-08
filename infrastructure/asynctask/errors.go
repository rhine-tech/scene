package asynctask

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(1, "asynctask")

var (
	ErrInternal            = _eg.CreateError(1, "internal error")
	ErrTaskNotFound        = _eg.CreateError(2, "task not found")
	ErrInvalidQueueTask    = _eg.CreateError(3, "invalid queue task")
	ErrTaskHandlerNotFound = _eg.CreateError(4, "task handler not found")
	ErrQueueConfigConflict = _eg.CreateError(5, "task queue config conflict")
	ErrInvalidQueueName    = _eg.CreateError(6, "invalid queue name")
	ErrQueueNotRegistered  = _eg.CreateError(7, "task queue not registered")
	ErrInvalidTaskType     = _eg.CreateError(8, "invalid task type")
	ErrInvalidTaskHandler  = _eg.CreateError(9, "invalid task handler")
	ErrInvalidQueuePayload = _eg.CreateError(10, "invalid queue payload")
)

package asynctask

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(5, "asynctask")

var (
	ErrInternal     = _eg.CreateError(1, "internal error")
	ErrTaskNotFound = _eg.CreateError(2, "task not found")
)

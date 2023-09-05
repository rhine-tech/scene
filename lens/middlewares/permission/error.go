package permission

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(1, "permission")

var (
	ErrPermissionDenied        = _eg.CreateError(0, "permission denied")
	ErrPermissionNotFound      = _eg.CreateError(1, "permission not found")
	ErrPermissionAlreadyExists = _eg.CreateError(2, "permission already exists")
)

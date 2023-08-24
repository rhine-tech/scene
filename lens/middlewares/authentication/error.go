package authentication

import "github.com/aynakeya/scene/errcode"

var _eg = errcode.NewErrorGroup(2, "authentication")
var (
	ErrAuthenticationFailed = _eg.CreateError(0, "authentication failed")
	ErrUserAlreadyExists    = _eg.CreateError(1, "user already exists")
	ErrUserNotFound         = _eg.CreateError(2, "user not found")
	ErrNotLogin             = _eg.CreateError(3, "not login")
)

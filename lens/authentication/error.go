package authentication

import "github.com/rhine-tech/scene/errcode"

var _eg = errcode.NewErrorGroup(2, "authentication")
var (
	ErrAuthenticationFailed = _eg.CreateError(0, "authentication failed")
	ErrUserAlreadyExists    = _eg.CreateError(1, "user already exists")
	ErrUserNotFound         = _eg.CreateError(2, "user not found")
	ErrNotLogin             = _eg.CreateError(3, "not login")
	ErrTokenNotFound        = _eg.CreateError(4, "token not found")
	ErrFailToAddUser        = _eg.CreateError(5, "fail to add user")
	ErrInternalError        = _eg.CreateError(6, "internal error")
	ErrTokenExpired         = _eg.CreateError(7, "token expired")
	ErrFailToGetToken       = _eg.CreateError(8, "fail to get token")
)

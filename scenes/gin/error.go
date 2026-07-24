package gin

import "github.com/rhine-tech/scene/errcode"

var (
	// ErrAlreadyDone tells Handle that the action already wrote the response.
	ErrAlreadyDone = errcode.CreateError(100, "gin already done")
)

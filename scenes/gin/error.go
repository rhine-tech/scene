package gin

import "github.com/rhine-tech/scene/errcode"

var (
	ErrAlreadyDone = errcode.CreateError(100, "gin already done")
)

package errcode

var (
	_eg                 = NewErrorGroup(0, "common")
	Success             = _eg.CreateError(0, "success")
	UnknownError        = _eg.CreateError(1, "unknown")
	ParameterError      = _eg.CreateError(2, "parameter error")
	InternalError       = _eg.CreateError(3, "internal error")
	RepositoryInitError = _eg.CreateError(4, "repository init error")
)

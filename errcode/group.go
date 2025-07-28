package errcode

import "fmt"

type ErrorGroup struct {
	Code int
	Name string
}

var _groups = map[int]string{}

func NewErrorGroup(groupCode int, name string) *ErrorGroup {
	if groupName, ok := _groups[groupCode]; ok {
		panic(fmt.Sprintf("error group %d(%s) already exists, please choose another one", groupCode, groupName))
	}
	eg := &ErrorGroup{
		Code: groupCode,
		Name: name,
	}
	_groups[groupCode] = name
	return eg
}

func (e *ErrorGroup) CreateError(code int, message string) *Error {
	return CreateError(e.Code*errcodeMask+code, message)
}

func (e *ErrorGroup) IsInGroup(err error) bool {
	if err == nil {
		return false
	}
	if e2, ok := err.(*Error); ok {
		return (e2.Code / errcodeMask) == e.Code
	}
	return false
}

func WithErrGroup(groupCode int) *ErrorGroup {
	if name, ok := _groups[groupCode]; ok {
		return &ErrorGroup{
			Code: groupCode,
			Name: name,
		}
	}
	return NewErrorGroup(groupCode, fmt.Sprintf("group_%d", groupCode))
}

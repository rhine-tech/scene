package errcode

import (
	"errors"
	"fmt"
	"strconv"
)

const errcodeMask = 1000

// errcode format    [err group] [sub err]
// 				   [...any digi] [3 digits]

// Error struct for pb.Error
type Error struct {
	Code    int
	Message string
	Detail  string
}

func (e *Error) Is(target error) bool {
	er, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == er.Code
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("err%d", e.Code)), nil
}

func (e *Error) UnmarshalJSON(b []byte) (err error) {
	if len(b) < 4 {
		return errors.New("invalid error code")
	}
	code, err := strconv.Atoi(string(b[3:]))
	if err != nil {
		return err
	}
	if x, ok := _codes[code]; ok {
		e.Code = code
		e.Message = x
		return nil
	}
	return errors.New("invalid error code")
}

var _codes = map[int]string{}

func CreateError(code int, msg string) *Error {
	if _, ok := _codes[code]; ok {
		panic(fmt.Sprintf("error %d already exists, please choose another code", code))
	}
	_codes[code] = msg
	return &Error{Code: code, Message: msg}
}

func (e *Error) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("Error %d: %s (%s)", e.Code, e.Message, e.Detail)
}

func (e *Error) WithDetail(err error) *Error {
	return &Error{
		e.Code,
		e.Message,
		err.Error(),
	}
}

func (e *Error) WithDetailStr(err string) *Error {
	return &Error{
		e.Code,
		e.Message,
		err,
	}
}

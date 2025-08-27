package errcode

import (
	"encoding/json"
	"errors"
)

type UnmarshalError struct {
	Error error
}

func (ue *UnmarshalError) UnmarshalJSON(data []byte) error {
	var s struct {
		Error *Error
	}
	if err := json.Unmarshal(data, &s); err == nil {
		if s.Error == nil {
			ue.Error = nil
			return nil
		}
		ue.Error = s.Error
		return nil
	}
	var e struct {
		Error string
	}
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}
	ue.Error = errors.New(e.Error)
	return nil
}

package errcode

import (
	"encoding/json"
	"errors"
)

type UnmarshalError struct {
	Error error
}

func (ue UnmarshalError) MarshalJSON() ([]byte, error) {
	if ue.Error == nil {
		return json.Marshal(struct {
			Error *Error
		}{})
	}
	if e, ok := ue.Error.(*Error); ok {
		return json.Marshal(struct {
			Error *Error
		}{Error: e})
	}
	return json.Marshal(struct {
		Error string
	}{Error: ue.Error.Error()})
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

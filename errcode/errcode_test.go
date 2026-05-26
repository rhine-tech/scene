package errcode

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestError_Is(t *testing.T) {
	require.True(t, UnknownError.WithDetailStr("asdf").Is(UnknownError))
	require.True(t, UnknownError.WithDetailStr("asdf").Is(UnknownError.WithDetailStr("asdf")))
	require.False(t, UnknownError.Is(Success))
}

func TestError_ErrorsIs(t *testing.T) {
	require.True(t, errors.Is(UnknownError.WithDetailStr("asdf"), UnknownError))
	require.True(t, errors.Is(UnknownError.WithDetailStr("asdf"), UnknownError.WithDetailStr("asdf")))
	require.False(t, errors.Is(UnknownError, Success))
}

func TestError_MarshalJSON(t *testing.T) {
	val, err := json.Marshal(UnknownError.WithDetailStr("asdf"))
	require.NoError(t, err)
	var e Error
	err = json.Unmarshal(val, &e)
	require.NoError(t, err)
	require.Equal(t, UnknownError.Code, e.Code)
}

type testE struct {
	Error1 UnmarshalError
}

func TestUnmarshalError_MarshalJSON_GenericError(t *testing.T) {
	val := testE{Error1: UnmarshalError{Error: errors.New("err1")}}
	data, err := json.Marshal(val)
	require.NoError(t, err)
	var val2 testE
	err = json.Unmarshal(data, &val2)
	require.NoError(t, err)
	require.EqualError(t, val2.Error1.Error, "err1")

}

func TestUnmarshalError_MarshalJSON_Errcode(t *testing.T) {
	val := testE{Error1: UnmarshalError{Error: ParameterError.WithDetailStr("bad id")}}
	data, err := json.Marshal(val)
	require.NoError(t, err)
	var val2 testE
	err = json.Unmarshal(data, &val2)
	require.NoError(t, err)
	require.True(t, errors.Is(val2.Error1.Error, ParameterError))
}

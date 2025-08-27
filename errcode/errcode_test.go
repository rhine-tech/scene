package errcode

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Error1 error
}

func TestAA(t *testing.T) {
	val := testE{Error1: errors.New("err1")}
	data, err := json.Marshal(val)
	require.NoError(t, err)
	fmt.Println(string(data))
	//data = []byte("{\"error1\":\"123\"}")
	var val2 testE
	err = json.Unmarshal(data, &val2)
	require.NoError(t, err)
	fmt.Println(val2)
}

package errcode

import (
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

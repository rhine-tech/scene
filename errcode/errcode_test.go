package errcode

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestError_Is(t *testing.T) {
	require.True(t, UnknownError.WithDetailStr("asdf").Is(UnknownError))
	require.True(t, UnknownError.WithDetailStr("asdf").Is(UnknownError.WithDetailStr("asdf")))
	require.False(t, UnknownError.Is(Success))
}

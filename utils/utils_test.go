package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsValidAddress(t *testing.T) {
	require.True(t, IsValidAddress("0.0.0.0:9090"))
	require.True(t, IsValidAddress(":9090"))
}

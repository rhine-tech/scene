package scene

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type testInterface interface {
}

func TestGetInterfaceName(t *testing.T) {
	require.Equal(t, "scene.testInterface", GetInterfaceName[testInterface]())
}

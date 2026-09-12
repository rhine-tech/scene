package gin

import (
	"net"
	"testing"

	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

func TestGinContainerStartReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	container, err := (Factory{
		Addr:    listener.Addr().String(),
		Routers: []RouterDefinition{DefaultRouter("/")},
	}).Build(registry.NewScope(), nil)
	require.NoError(t, err)
	t.Cleanup(container.(*ginContainer).cancel)

	err = container.Start()
	require.Error(t, err)
	var opErr *net.OpError
	require.ErrorAs(t, err, &opErr)
	require.Equal(t, "listen", opErr.Op)
}

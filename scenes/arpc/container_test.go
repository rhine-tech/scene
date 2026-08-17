package arpc

import (
	"context"
	"net"
	"testing"

	libarpc "github.com/lesismal/arpc"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/stretchr/testify/require"
)

func TestARPCContainerStartReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	container := &arpcContainer{
		addr:   listener.Addr().String(),
		server: libarpc.NewServer(),
		log:    logger.NoopLogger{},
	}

	err = container.Start()
	require.Error(t, err)
	var opErr *net.OpError
	require.ErrorAs(t, err, &opErr)
	require.Equal(t, "listen", opErr.Op)
}

func TestARPCContainerStartAndStop(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())
	container := &arpcContainer{
		addr:   addr,
		server: libarpc.NewServer(),
		log:    logger.NoopLogger{},
	}

	require.NoError(t, container.Start())
	require.NoError(t, container.Stop(context.Background()))
}

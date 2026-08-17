package websocket

import (
	"context"
	"net"
	"testing"

	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/stretchr/testify/require"
)

func TestWebSocketContainerStartReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	container := &websocketContainer{
		addr:   listener.Addr().String(),
		mux:    NewWebSocketMux(),
		logger: logger.NoopLogger{},
	}

	err = container.Start()
	require.Error(t, err)
	var opErr *net.OpError
	require.ErrorAs(t, err, &opErr)
	require.Equal(t, "listen", opErr.Op)
}

func TestWebSocketContainerStartAndStop(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())
	container := &websocketContainer{
		addr:   addr,
		mux:    NewWebSocketMux(),
		logger: logger.NoopLogger{},
	}

	require.NoError(t, container.Start())
	require.NoError(t, container.Stop(context.Background()))
}

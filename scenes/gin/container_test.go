package gin

import (
	"context"
	"net"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/stretchr/testify/require"
)

func TestGinContainerStartReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	container := &ginContainer{
		addr:    listener.Addr().String(),
		engine:  gin.New(),
		logger:  logger.NoopLogger{},
		baseCtx: context.Background(),
	}

	err = container.Start()
	require.Error(t, err)
	var opErr *net.OpError
	require.ErrorAs(t, err, &opErr)
	require.Equal(t, "listen", opErr.Op)
}

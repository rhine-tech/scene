package mcp

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

func newTestLoader(t *testing.T) *scene.ModuleLoader {
	t.Helper()
	loader := scene.NewModuleLoaderIn(registry.NewScope(), nil)
	require.NoError(t, loader.Init())
	t.Cleanup(func() { require.NoError(t, loader.TearDown()) })
	return loader
}

func TestStreamableHTTPTransportOwnsRoutingAndServer(t *testing.T) {
	loader := newTestLoader(t)
	definition := NewStreamableHTTPFactory(
		"test",
		"v1",
		"127.0.0.1:8080",
		[]StreamableHTTPOption{
			WithStreamableHTTPEndpointPath("/custom"),
			WithStreamableHTTPServer(func(handler http.Handler) *http.Server {
				return &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.Header().Set("X-Test-Server", "custom")
					handler.ServeHTTP(writer, request)
				})}
			}),
		},
		nil,
	)
	built, err := definition.Build(loader)
	require.NoError(t, err)
	handler := built.(*StreamableHTTPScene).httpServer.Handler

	notFound := httptest.NewRecorder()
	handler.ServeHTTP(notFound, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{}")))
	require.Equal(t, http.StatusNotFound, notFound.Code)
	require.Equal(t, "custom", notFound.Header().Get("X-Test-Server"))

	configured := httptest.NewRecorder()
	handler.ServeHTTP(configured, httptest.NewRequest(http.MethodPost, "/custom", strings.NewReader("{}")))
	require.NotEqual(t, http.StatusNotFound, configured.Code)
}

func TestSSETransportUsesConfiguredHTTPServer(t *testing.T) {
	loader := newTestLoader(t)
	definition := NewSSEFactory(
		"test",
		"v1",
		"127.0.0.1:8080",
		[]SSEOption{WithSSEHTTPServer(func(handler http.Handler) *http.Server {
			return &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("X-Test-Server", "custom")
				handler.ServeHTTP(writer, request)
			})}
		})},
		nil,
	)
	built, err := definition.Build(loader)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	built.(*SSEScene).httpServer.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, "custom", response.Header().Get("X-Test-Server"))
}

func TestStreamableHTTPTransportValidatesTLSBeforeListening(t *testing.T) {
	loader := newTestLoader(t)
	missingCert := filepath.Join(t.TempDir(), "missing-cert.pem")
	definition := NewStreamableHTTPFactory(
		"test",
		"v1",
		"127.0.0.1:18080",
		[]StreamableHTTPOption{WithStreamableHTTPTLS(missingCert, missingCert+".key")},
		nil,
	)
	built, err := definition.Build(loader)
	require.NoError(t, err)

	err = built.Start()
	var pathError *os.PathError
	require.ErrorAs(t, err, &pathError)
}

func TestMCPScenesStartReturnListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	loader := newTestLoader(t)
	definitions := []scene.SceneDefinition{
		NewSSEFactory("test", "v1", listener.Addr().String(), nil, nil),
		NewStreamableHTTPFactory("test", "v1", listener.Addr().String(), nil, nil),
	}
	for _, definition := range definitions {
		built, buildErr := definition.Build(loader)
		require.NoError(t, buildErr)

		startErr := built.Start()
		require.Error(t, startErr)
		var opErr *net.OpError
		require.ErrorAs(t, startErr, &opErr)
		require.Equal(t, "listen", opErr.Op)
	}
}

func TestMCPScenesStartAndStop(t *testing.T) {
	loader := newTestLoader(t)
	factories := []func(string) scene.SceneDefinition{
		func(addr string) scene.SceneDefinition { return NewSSEFactory("test", "v1", addr, nil, nil) },
		func(addr string) scene.SceneDefinition {
			return NewStreamableHTTPFactory("test", "v1", addr, nil, nil)
		},
	}
	for _, factory := range factories {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := listener.Addr().String()
		require.NoError(t, listener.Close())

		definition := factory(addr)
		built, err := definition.Build(loader)
		require.NoError(t, err)
		require.NoError(t, built.Start())
		require.NoError(t, built.Stop(context.Background()))
	}
}

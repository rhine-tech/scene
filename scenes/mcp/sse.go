package mcp

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
)

type SSEScene struct {
	addr       string
	server     *server.MCPServer
	sse        *server.SSEServer
	httpServer *http.Server
	apps       []McpApp
	logger     logger.ILogger
}

// SSEFactory builds an MCP SSE scene from MCP applications.
type SSEFactory struct {
	Name          string
	Version       string
	Addr          string
	SSEOptions    []SSEOption
	ServerOptions []ServerOption
}

var _ scene.SceneFactory[McpApp] = SSEFactory{}

// NewSSEFactory declares an MCP SSE scene.
func NewSSEFactory(
	name string,
	version string,
	addr string,
	sseOptions []SSEOption,
	serverOptions []ServerOption,
) scene.SceneDefinition {
	return scene.WithScene[McpApp](SSEFactory{
		Name:          name,
		Version:       version,
		Addr:          addr,
		SSEOptions:    sseOptions,
		ServerOptions: serverOptions,
	})
}

func (f SSEFactory) Build(scope *registry.Scope, apps []McpApp) (scene.Scene, error) {
	mcpServer := server.NewMCPServer(f.Name, f.Version, f.ServerOptions...)
	options := resolveSSEOptions(f.SSEOptions)
	sseServer := server.NewSSEServer(mcpServer, options.handlerOptions...)
	httpServer, err := buildHTTPServer(f.Addr, sseServer, options.httpServerBuilder)
	if err != nil {
		return nil, err
	}
	log, exists := registry.LookupIn[logger.ILogger](scope)
	if !exists {
		log = logger.NoopLogger{}
	}
	return &SSEScene{
		server:     mcpServer,
		sse:        sseServer,
		httpServer: httpServer,
		addr:       f.Addr,
		apps:       apps,
		logger:     log.WithPrefix((&SSEScene{}).ImplName().Identifier()),
	}, nil
}

func (m *SSEScene) ImplName() scene.ImplName {
	return scene.NewSceneImplName("mcp", "Scene", "sse")
}

func (m *SSEScene) Start() error {
	if !utils.IsValidAddress(m.addr) {
		m.logger.Errorf("invalid address: %s", m.addr)
		return errors.New("invalid address " + m.addr)
	}
	for _, app := range m.apps {
		if err := app.Register(m.server); err != nil {
			return err
		}
	}
	listener, err := net.Listen("tcp", m.addr)
	if err != nil {
		return err
	}
	m.logger.Infof("mcp sse server started, listen on 'http://%s'", utils.PrettyAddress(m.addr))
	go func() {
		if err := m.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Errorf("listen failed: %s\n", err)
		}
	}()
	return nil
}

func (m *SSEScene) Stop(ctx context.Context) error {
	m.sse.CloseSessions()
	return m.httpServer.Shutdown(ctx)
}

func (m *SSEScene) ListAppNames() []string {
	return nil
}

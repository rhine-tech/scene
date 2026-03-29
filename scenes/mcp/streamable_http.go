package mcp

import (
	"context"
	"errors"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
)

type StreamableHTTPScene struct {
	addr   string
	server *server.MCPServer
	http   *server.StreamableHTTPServer
	clos   func(ctx context.Context) error
	apps   []McpApp
	logger logger.ILogger `aperture:""`
}

func (m *StreamableHTTPScene) ImplName() scene.ImplName {
	return scene.NewSceneImplName("mcp", "Scene", "streamable-http")
}

func (m *StreamableHTTPScene) Start() error {
	if !utils.IsValidAddress(m.addr) {
		registry.Logger.Errorf("invalid address: %s", m.addr)
		return errors.New("invalid address " + m.addr)
	}
	for _, app := range m.apps {
		if err := app.Register(m.server); err != nil {
			return err
		}
	}
	m.clos = func(ctx context.Context) error {
		return m.http.Shutdown(ctx)
	}
	go func() {
		m.logger.Infof("mcp streamable http server started, listen on 'http://%s'", utils.PrettyAddress(m.addr))
		if err := m.http.Start(m.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Errorf("listen failed: %s\n", err)
		}
	}()
	return nil
}

func (m *StreamableHTTPScene) Stop(ctx context.Context) error {
	return m.clos(ctx)
}

func (m *StreamableHTTPScene) ListAppNames() []string {
	names := make([]string, 0, len(m.apps))
	for _, app := range m.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}

func NewStreamableHTTP(
	name string,
	version string,
	addr string,
	apps []McpApp,
	httpOpts []StreamableHTTPOption,
	serverOpts []ServerOption,
) scene.Scene {
	s := server.NewMCPServer(name, version, serverOpts...)
	httpServer := server.NewStreamableHTTPServer(s, httpOpts...)
	return &StreamableHTTPScene{
		server: s,
		http:   httpServer,
		addr:   addr,
		apps:   apps,
		clos:   func(ctx context.Context) error { return httpServer.Shutdown(ctx) },
		logger: registry.Logger.WithPrefix((&StreamableHTTPScene{}).ImplName().Identifier()),
	}
}

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

type SSEScene struct {
	addr   string
	server *server.MCPServer
	sse    *server.SSEServer
	clos   func(ctx context.Context) error
	apps   []McpApp
	logger logger.ILogger `aperture:""`
}

func (m *SSEScene) ImplName() scene.ImplName {
	return scene.NewSceneImplName("mcp", "Scene", "sse")
}

func (m *SSEScene) Start() error {
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
		return m.sse.Shutdown(ctx)
	}
	go func() {
		m.logger.Infof("mcp sse server started, listen on 'http://%s'", utils.PrettyAddress(m.addr))
		if err := m.sse.Start(m.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Errorf("listen failed: %s\n", err)
		}
	}()
	return nil
}

func (m *SSEScene) Stop(ctx context.Context) error {
	return m.clos(ctx)
}

func (m *SSEScene) ListAppNames() []string {
	return nil
}

func NewSSE(
	name string,
	version string,
	addr string,
	apps []McpApp,
	sseOpts []SSEOption,
	serverOpts []ServerOption) scene.Scene {
	s := server.NewMCPServer(
		name, version, serverOpts...,
	)
	sseServer := server.NewSSEServer(s, sseOpts...)
	return &SSEScene{
		server: s,
		sse:    sseServer,
		addr:   addr,
		apps:   apps,
		clos:   func(ctx context.Context) error { return sseServer.Shutdown(ctx) },
		logger: registry.Logger.WithPrefix((&SSEScene{}).ImplName().Identifier()),
	}
}

package mcp

import (
	"context"
	"errors"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
	"net/http"
)

type McpScene struct {
	addr   string
	server *server.MCPServer
	clos   func(ctx context.Context) error
	apps   []McpApp
	logger logger.ILogger `aperture:""`
}

func (m *McpScene) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("mcp", "Scene")
}

func (m *McpScene) Start() error {
	if !utils.IsValidAddress(m.addr) {
		registry.Logger.Errorf("invalid address: %s", m.addr)
		return errors.New("invalid address " + m.addr)
	}
	for _, app := range m.apps {
		if err := app.Register(m.server); err != nil {
			return err
		}
	}
	// todo: options
	srv := server.NewSSEServer(m.server)
	m.clos = func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}
	go func() {
		m.logger.Infof("mcp sse server started, listen on 'http://%s'", utils.PrettyAddress(m.addr))
		if err := srv.Start(m.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Errorf("listen failed: %s\n", err)
		}
	}()
	return nil
}

func (m *McpScene) Stop(ctx context.Context) error {
	return m.clos(ctx)
}

func (m *McpScene) ListAppNames() []string {
	return nil
}

func NewMcpScene(
	name string,
	version string,
	addr string,
	apps []McpApp) scene.Scene {
	s := server.NewMCPServer(
		name, version,
		// todo: mcp server options
	)
	return &McpScene{
		server: s,
		addr:   addr,
		apps:   apps,
		clos:   func(ctx context.Context) error { return nil },
		logger: registry.Logger.WithPrefix((&McpScene{}).ImplName().Identifier()),
	}
}

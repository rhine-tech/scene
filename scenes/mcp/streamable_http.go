package mcp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
)

type StreamableHTTPScene struct {
	addr        string
	server      *server.MCPServer
	handler     *server.StreamableHTTPServer
	httpServer  *http.Server
	apps        []McpApp
	logger      logger.ILogger
	tlsCertFile string
	tlsKeyFile  string
}

// StreamableHTTPFactory builds an MCP Streamable HTTP scene from MCP applications.
type StreamableHTTPFactory struct {
	Name          string
	Version       string
	Addr          string
	HTTPOptions   []StreamableHTTPOption
	ServerOptions []ServerOption
}

var _ scene.SceneFactory[McpApp] = StreamableHTTPFactory{}

// NewStreamableHTTPFactory declares an MCP Streamable HTTP scene.
func NewStreamableHTTPFactory(
	name string,
	version string,
	addr string,
	httpOptions []StreamableHTTPOption,
	serverOptions []ServerOption,
) scene.SceneDefinition {
	return scene.WithScene[McpApp](StreamableHTTPFactory{
		Name:          name,
		Version:       version,
		Addr:          addr,
		HTTPOptions:   httpOptions,
		ServerOptions: serverOptions,
	})
}

func (f StreamableHTTPFactory) Build(scope *registry.Scope, apps []McpApp) (scene.Scene, error) {
	mcpServer := server.NewMCPServer(f.Name, f.Version, f.ServerOptions...)
	options := resolveStreamableHTTPOptions(f.HTTPOptions)
	handler := server.NewStreamableHTTPServer(mcpServer, options.handlerOptions...)
	routedHandler := routeStreamableHTTP(options.endpointPath, handler)
	httpServer, err := buildHTTPServer(f.Addr, routedHandler, options.httpServerBuilder)
	if err != nil {
		return nil, err
	}
	log, exists := registry.LookupIn[logger.ILogger](scope)
	if !exists {
		log = logger.NoopLogger{}
	}
	return &StreamableHTTPScene{
		server:      mcpServer,
		handler:     handler,
		httpServer:  httpServer,
		addr:        f.Addr,
		apps:        apps,
		logger:      log.WithPrefix((&StreamableHTTPScene{}).ImplName().Identifier()),
		tlsCertFile: options.tlsCertFile,
		tlsKeyFile:  options.tlsKeyFile,
	}, nil
}

func (m *StreamableHTTPScene) ImplName() scene.ImplName {
	return scene.NewSceneImplName("mcp", "Scene", "streamable-http")
}

func (m *StreamableHTTPScene) Start() error {
	if !utils.IsValidAddress(m.addr) {
		m.logger.Errorf("invalid address: %s", m.addr)
		return errors.New("invalid address " + m.addr)
	}
	for _, app := range m.apps {
		if err := app.Register(m.server); err != nil {
			return err
		}
	}
	if (m.tlsCertFile == "") != (m.tlsKeyFile == "") {
		return errors.New("both TLS cert and key must be provided")
	}
	var tlsConfig *tls.Config
	if m.tlsCertFile != "" {
		certificate, err := tls.LoadX509KeyPair(m.tlsCertFile, m.tlsKeyFile)
		if err != nil {
			return err
		}
		if m.httpServer.TLSConfig == nil {
			tlsConfig = new(tls.Config)
		} else {
			tlsConfig = m.httpServer.TLSConfig.Clone()
		}
		tlsConfig.Certificates = append([]tls.Certificate{certificate}, tlsConfig.Certificates...)
		m.httpServer.TLSConfig = tlsConfig
	}
	listener, err := net.Listen("tcp", m.addr)
	if err != nil {
		return err
	}
	scheme := "http"
	if tlsConfig != nil {
		scheme = "https"
	}
	m.logger.Infof("mcp streamable http server started, listen on '%s://%s'", scheme, utils.PrettyAddress(m.addr))
	go func() {
		var err error
		if tlsConfig == nil {
			err = m.httpServer.Serve(listener)
		} else {
			err = m.httpServer.ServeTLS(listener, "", "")
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.logger.Errorf("listen failed: %s\n", err)
		}
	}()
	return nil
}

func (m *StreamableHTTPScene) Stop(ctx context.Context) error {
	return errors.Join(m.handler.Shutdown(ctx), m.httpServer.Shutdown(ctx))
}

func (m *StreamableHTTPScene) ListAppNames() []string {
	names := make([]string, 0, len(m.apps))
	for _, app := range m.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}

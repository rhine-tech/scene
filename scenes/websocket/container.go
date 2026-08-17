package websocket

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
)

type websocketContainer struct {
	addr   string
	prefix string
	mux    *WebSocketMux
	apps   []WebsocketApplication
	logger logger.ILogger
	server *http.Server
}

// Factory builds a WebSocket scene from WebSocket applications.
type Factory struct {
	Addr    string
	Options []WsOption
}

var _ scene.SceneFactory[WebsocketApplication] = Factory{}

// NewFactory declares a WebSocket scene.
func NewFactory(addr string, options ...WsOption) scene.SceneDefinition {
	return scene.WithScene[WebsocketApplication](Factory{
		Addr:    addr,
		Options: options,
	})
}

func (f Factory) Build(scope *registry.Scope, apps []WebsocketApplication) (scene.Scene, error) {
	wsMux := NewWebSocketMux()
	for _, option := range f.Options {
		if err := option(scope, wsMux); err != nil {
			return nil, err
		}
	}
	container := &websocketContainer{
		addr: f.Addr,
		apps: apps,
		mux:  wsMux,
	}
	log, exists := registry.LookupIn[logger.ILogger](scope)
	if !exists {
		log = logger.NoopLogger{}
	}
	container.logger = log.WithPrefix(container.ImplName().Identifier())
	return container, nil
}

func (g *websocketContainer) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("websocket", "Scene")
}

func (c *websocketContainer) startApps() error {
	created := 0
	for _, app := range c.apps {
		prefix := app.Prefix()
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}
		if err := app.Create(c.mux.UsePrefix(prefix)); err != nil {
			c.logger.Errorf("failed to create %s: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("%s created", app.Name())
			created++
		}
	}
	c.logger.Infof("created %d apps, failed to create %d app", created, len(c.apps)-created)
	endpoints := ""
	endpointsCount := 0
	_ = c.mux.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if route.GetHandler() == nil {
			return nil
		}
		t, _ := route.GetPathTemplate()
		endpoints += fmt.Sprintf("%8s %s\n", "-", t)
		endpointsCount++
		return nil
	})
	c.logger.Infof("registered %d endpoint\n\n%s", endpointsCount, endpoints)
	return nil
}

func (c *websocketContainer) stopApps() error {
	for _, app := range c.apps {
		if err := app.Destroy(); err != nil {
			c.logger.Errorf("%s failed to destroy: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("%s destroyed", app.Name())
		}
	}
	return nil
}

func (c *websocketContainer) Start() error {
	if !utils.IsValidAddress(c.addr) {
		c.logger.Errorf("invalid address: %s", c.addr)
		return errors.New("invalid address " + c.addr)
	}
	if err := c.startApps(); err != nil {
		return err
	}
	c.server = &http.Server{
		Addr:    c.addr,
		Handler: c.mux,
	}
	listener, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	c.logger.Infof("websocket server started, listen on 'ws://%s'", utils.PrettyAddress(c.addr))
	go func() {
		if err := c.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Errorf("listen: %s\n", err)
		}
	}()
	return nil
}

func (c *websocketContainer) Stop(ctx context.Context) error {

	if err := c.stopApps(); err != nil {
		return err
	}

	subctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.server.Shutdown(subctx); err != nil {
		c.logger.Infof("Server Shutdown:", err)
		return err
	}
	return nil
}

func (g *websocketContainer) ListAppNames() []string {
	names := make([]string, 0, len(g.apps))
	for _, app := range g.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}

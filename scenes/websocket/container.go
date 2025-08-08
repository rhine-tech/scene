package websocket

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
	"net/http"
	"strings"
	"time"
)

type websocketContainer struct {
	addr   string
	prefix string
	mux    *WebSocketMux
	apps   []WebsocketApplication
	logger logger.ILogger
	server *http.Server
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
		registry.Logger.Errorf("invalid address: %s", c.addr)
		return errors.New("invalid address " + c.addr)
	}
	if err := c.startApps(); err != nil {
		return err
	}
	c.server = &http.Server{
		Addr:    c.addr,
		Handler: c.mux,
	}
	go func() {
		c.logger.Infof("websocket server started, listen on 'ws://%s'", utils.PrettyAddress(c.addr))
		if err := c.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func NewContainer(addr string, apps []WebsocketApplication, opts ...WsOption) scene.Scene {
	wsMux := NewWebSocketMux()
	for _, opt := range opts {
		_ = opt(wsMux)
	}
	container := &websocketContainer{
		addr: addr,
		apps: apps,
		mux:  wsMux,
	}
	container.logger = registry.Logger.WithPrefix(container.ImplName().Identifier())
	return container
}

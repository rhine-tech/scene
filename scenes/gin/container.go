package gin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
)

func createGinEngine(scope *registry.Scope) *gin.Engine {
	cfg, exists := registry.LookupIn[config.IConfig](scope)
	if exists && cfg.GetBool("scene.debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	return engine
}

type ginContainer struct {
	addr    string
	routers prefixMux
	handler http.Handler
	apps    []routedApplication
	logger  logger.ILogger
	server  *http.Server
	baseCtx context.Context
	cancel  context.CancelFunc
}

type routedApplication struct {
	GinApplication
	router *ginRouter
}

// Factory builds a Gin scene from Gin applications.
type Factory struct {
	Addr            string
	Routers         []RouterDefinition
	HTTPMiddlewares []HTTPMiddleware
}

var _ scene.SceneFactory[GinApplication] = Factory{}

// NewFactory declares a Gin scene.
func NewFactory(addr string, routers ...RouterDefinition) scene.SceneDefinition {
	return scene.WithScene[GinApplication](Factory{
		Addr:    addr,
		Routers: routers,
	})
}

func (f Factory) Build(scope *registry.Scope, apps []GinApplication) (scene.Scene, error) {
	container := &ginContainer{addr: f.Addr}
	byName := make(map[string]*ginRouter, len(f.Routers))
	prefixes := make(map[string]string, len(f.Routers))
	for _, definition := range f.Routers {
		if _, exists := byName[definition.Name]; exists {
			return nil, fmt.Errorf("scene-gin: duplicate router name %q", definition.Name)
		}
		prefix := path.Clean("/" + definition.Prefix)
		if previous, exists := prefixes[prefix]; exists {
			return nil, fmt.Errorf("scene-gin: routers %q and %q use the same prefix %q", previous, definition.Name, prefix)
		}
		engine := createGinEngine(scope)
		for _, option := range definition.Options {
			if err := option(scope, engine); err != nil {
				return nil, fmt.Errorf("scene-gin: configure router %q: %w", definition.Name, err)
			}
		}
		router := &ginRouter{name: definition.Name, prefix: prefix, engine: engine}
		byName[definition.Name] = router
		prefixes[prefix] = definition.Name
		container.routers = append(container.routers, router)
	}
	for _, app := range apps {
		name := ""
		if selector, ok := app.(RouterSelector); ok {
			name = selector.RouterName()
		}
		router, exists := byName[name]
		if !exists {
			return nil, fmt.Errorf("scene-gin: application %s selects unconfigured router %q", app.Name(), name)
		}
		container.apps = append(container.apps, routedApplication{GinApplication: app, router: router})
	}
	slices.SortStableFunc(container.routers, func(a, b *ginRouter) int {
		return len(b.prefix) - len(a.prefix)
	})
	container.handler = container.routers
	// Wrap in reverse so requests enter middleware in declaration order.
	for i := len(f.HTTPMiddlewares) - 1; i >= 0; i-- {
		handler, err := f.HTTPMiddlewares[i](scope, container.handler)
		if err != nil {
			return nil, fmt.Errorf("scene-gin: configure HTTP middleware %d: %w", i, err)
		}
		container.handler = handler
	}
	container.baseCtx, container.cancel = context.WithCancel(context.Background())
	log, exists := registry.LookupIn[logger.ILogger](scope)
	if !exists {
		log = logger.NoopLogger{}
	}
	container.logger = log.WithPrefix(container.ImplName().Identifier())
	return container, nil
}

func (c *ginContainer) ImplName() scene.ImplName {
	return scene.NewSceneImplNameNoVer("gin", "Scene")
}

func (c *ginContainer) startApps() error {
	created := 0
	for _, app := range c.apps {
		router := app.router.engine.Group(app.router.prefix).Group(app.Prefix())
		if err := app.Create(app.router.engine, router); err != nil {
			c.logger.Errorf("failed to create %s: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("%s created", app.Name())
			created++
		}
	}
	c.logger.Infof("created %d apps, failed to create %d app", created, len(c.apps)-created)
	endpoints := ""
	endpointsCount := 0
	for _, router := range c.routers {
		c.logger.Infof("router %q mounted at %q", router.name, router.prefix)
		for _, route := range router.engine.Routes() {
			endpoints += fmt.Sprintf("%8s %-8s %s\n", "-", route.Method, route.Path)
			endpointsCount++
		}
	}
	c.logger.Infof("registered %d endpoint\n\n%s", endpointsCount, endpoints)
	return nil
}

func (c *ginContainer) stopApps() error {
	for _, app := range c.apps {
		if err := app.Destroy(); err != nil {
			c.logger.Errorf("%s failed to destroy: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("%s destroyed", app.Name())
		}
	}
	return nil
}

func (c *ginContainer) Start() error {
	if !utils.IsValidAddress(c.addr) {
		c.logger.Errorf("invalid address: %s", c.addr)
		return errors.New("invalid address " + c.addr)
	}
	if err := c.startApps(); err != nil {
		return err
	}
	c.server = &http.Server{
		Addr:    c.addr,
		Handler: c.handler,
		BaseContext: func(listener net.Listener) context.Context {
			return c.baseCtx
		},
	}
	listener, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	c.logger.Infof("gin http server started, listen on 'http://%s'", utils.PrettyAddress(c.addr))
	go func() {
		if err := c.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Errorf("listen: %s\n", err)
		}
	}()
	return nil
}

func (c *ginContainer) Stop(ctx context.Context) error {

	if err := c.stopApps(); err != nil {
		return err
	}
	if c.cancel != nil {
		// Cancel the base context first so long-lived handlers (SSE/WebSocket-like loops)
		// can observe ctx.Done() and exit before graceful shutdown timeout.
		c.cancel()
	}

	subctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.server.Shutdown(subctx); err != nil {
		c.logger.Infof("Server Shutdown: %v", err)
		return err
	}
	return nil
}

func (g *ginContainer) ListAppNames() []string {
	names := make([]string, 0, len(g.apps))
	for _, app := range g.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}

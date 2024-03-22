package gin

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils"
	"net/http"
	"time"
)

func createGinEngine() *gin.Engine {
	if registry.Config.GetBool("scene.debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	return engine
}

type ginContainer struct {
	addr   string
	prefix string
	engine *gin.Engine
	apps   []GinApplication
	logger logger.ILogger
	server *http.Server
}

func (c *ginContainer) Name() scene.ImplName {
	return scene.NewSceneImplNameNoVer("gin", "container")
}

func (c *ginContainer) startApps() error {
	router := c.engine.Group(c.prefix)
	created := 0
	for _, app := range c.apps {
		if err := app.Create(c.engine, router.Group(app.Prefix())); err != nil {
			c.logger.Errorf("failed to create %s: %s", app.Name(), err.Error())
		} else {
			c.logger.Infof("%s created", app.Name())
			created++
		}
	}
	c.logger.Infof("created %d apps, failed to create %d app", created, len(c.apps)-created)
	endpoints := ""
	endpointsCount := 0
	for _, route := range c.engine.Routes() {
		endpoints += fmt.Sprintf("%8s %-8s %s\n", "-", route.Method, route.Path)
		endpointsCount++
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
		registry.Logger.Errorf("invalid address: %s", c.addr)
		return errors.New("invalid address " + c.addr)
	}
	if err := c.startApps(); err != nil {
		return err
	}
	c.server = &http.Server{
		Addr:    c.addr,
		Handler: c.engine,
	}
	go func() {
		c.logger.Infof("gin http server started, listen on 'http://%s'", utils.PrettyAddress(c.addr))
		if err := c.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.logger.Errorf("listen: %s\n", err)
		}
	}()
	return nil
}

func (c *ginContainer) Stop(ctx context.Context) error {

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

func (g *ginContainer) ListAppNames() []string {
	names := make([]string, 0, len(g.apps))
	for _, app := range g.apps {
		names = append(names, app.Name().Identifier())
	}
	return names
}

func NewAppContainerWithPrefix(
	addr string,
	prefix string,
	apps []GinApplication,
	options ...GinOption,
) scene.ApplicationContainer {
	ginEngine := createGinEngine()
	for _, opt := range options {
		if err := opt(ginEngine); err != nil {
			panic(err)
		}
	}
	container := &ginContainer{
		addr:   addr,
		prefix: prefix,
		engine: ginEngine,
		apps:   apps,
	}
	container.logger = registry.Logger.WithPrefix(container.Name().Identifier())
	return container
}

func NewAppContainer(
	addr string,
	apps []GinApplication,
	options ...GinOption,
) scene.ApplicationContainer {
	return NewAppContainerWithPrefix(addr, "/", apps, options...)
}

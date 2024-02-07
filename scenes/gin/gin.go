package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
)

const SceneName = "scene.app-container.http.gin"

type GinApplication interface {
	scene.Application
	Prefix() string
	Create(engine *gin.Engine, router gin.IRouter) error
	Destroy() error
}

type CommonApp struct {
	AppError  error
	AppStatus scene.AppStatus
	Logger    logger.ILogger
}

func (s *CommonApp) Status() scene.AppStatus {
	return s.AppStatus
}

func (s *CommonApp) Error() error {
	return s.AppError
}

func (s *CommonApp) Destroy() error {
	return nil
}

type ginAppFactory struct {
	engine *gin.Engine
}

func NewAppFactory(engine *gin.Engine) scene.ApplicationFactory[GinApplication] {
	return &ginAppFactory{
		engine: engine,
	}
}

func (g *ginAppFactory) Name() string {
	return "scene.factory.gin"
}

func (g *ginAppFactory) Create(t GinApplication) error {
	return t.(GinApplication).Create(g.engine, g.engine.Group(t.Prefix()))
}

func (g *ginAppFactory) Destroy(t GinApplication) error {
	return t.(GinApplication).Destroy()
}

type Context[T GinApplication] struct {
	*gin.Context
	App T
}

func (g *Context[T]) Get(key string) (value any, exists bool) {
	return g.Context.Get(key)
}

func (g *Context[T]) Set(key string, value any) {
	g.Context.Set(key, value)
}

func GetContext(c *gin.Context) scene.Context {
	return &Context[GinApplication]{Context: c, App: GinApplication(nil)}
}

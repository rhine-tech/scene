package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/registry"
)

const SceneName = "scene.app-container.http.gin"

type GinApplication interface {
	scene.Application
	Prefix() string
	Create(engine *gin.Engine, router gin.IRouter) error
	Destroy() error
}

type AppRoutes[T any] struct {
	AppName  scene.ImplName
	BasePath string
	Actions  []Action[*T]
	Context  T
}

func (a *AppRoutes[T]) Name() scene.ImplName {
	return a.AppName
}

func (a *AppRoutes[T]) Prefix() string {
	return a.BasePath
}

func (a *AppRoutes[T]) Create(engine *gin.Engine, router gin.IRouter) error {
	registry.Inject(&a.Context)
	approuter := NewAppRouter(&a.Context, router)
	approuter.HandleActions(a.Actions...)
	return nil
}

func (a *AppRoutes[T]) Destroy() error {
	return nil
}

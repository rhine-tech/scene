package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
)

// GinApplication mounts one module's HTTP routes into a Gin scene.
//
// Create receives the selected Router's Gin engine and a *gin.RouterGroup
// scoped by the Router prefix and Prefix. Most implementations should register
// routes on router. Registering on engine cannot escape the selected Router.
type GinApplication interface {
	scene.Application
	Prefix() string
	Create(engine *gin.Engine, router gin.IRouter) error
	Destroy() error
}

// RouterSelector optionally selects the named router an application belongs to.
// Applications without RouterSelector use DefaultRouter.
type RouterSelector interface {
	RouterName() string
}

// AppRoutes is the declarative GinApplication used by most modules.
//
// Context is injected by the module loader before Create is called.
// Middlewares apply to every action, before middleware supplied by an
// individual MiddlewareProvider.
type AppRoutes[T any] struct {
	AppName     scene.ImplName
	Router      string // Router name; leave unset to use DefaultRouter.
	BasePath    string
	Actions     []Action[*T]
	Context     T `aperture:"embed"`
	Middlewares gin.HandlersChain
}

func (a *AppRoutes[T]) Name() scene.ImplName {
	return a.AppName
}

func (a *AppRoutes[T]) Prefix() string {
	return a.BasePath
}

func (a *AppRoutes[T]) RouterName() string {
	return a.Router
}

func (a *AppRoutes[T]) Create(engine *gin.Engine, router gin.IRouter) error {
	approuter := NewAppRouter(&a.Context, router, a.Middlewares)
	approuter.HandleActions(a.Actions...)
	return nil
}

func (a *AppRoutes[T]) Destroy() error {
	return nil
}

package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"net/http"
)

type Action[T any] interface {
	Request[T]
	scene.HttpRoute
	Middleware() gin.HandlersChain
}

type BaseAction struct {
}

func (a *BaseAction) Middleware() gin.HandlersChain {
	return nil
}

type AppRouter[T any] struct {
	app    T
	router gin.IRouter
}

func NewAppRouter[T any](app T, router gin.IRouter) *AppRouter[T] {
	return &AppRouter[T]{app: app, router: router}
}

func (r *AppRouter[T]) Router() gin.IRouter {
	return r.router
}

func (r *AppRouter[T]) HandleAction(action Action[T]) {
	routeInfo := action.GetRoute()
	chain := MiddlewareChain(action.Middleware()...)
	switch routeInfo.Method {
	case http.MethodGet:
		r.router.GET(routeInfo.Path, chain(Handle(r.app, action))...)
	case http.MethodPost:
		r.router.POST(routeInfo.Path, chain(Handle(r.app, action))...)
	case http.MethodDelete:
		r.router.DELETE(routeInfo.Path, chain(Handle(r.app, action))...)
	case http.MethodHead:
		r.router.HEAD(routeInfo.Path, chain(Handle(r.app, action))...)
	case http.MethodPut:
		r.router.PUT(routeInfo.Path, chain(Handle(r.app, action))...)
	case http.MethodOptions:
		r.router.OPTIONS(routeInfo.Path, chain(Handle(r.app, action))...)
	default:
		panic("invalid method")
	}
}

func (r *AppRouter[T]) HandleActions(actions ...Action[T]) {
	for _, action := range actions {
		r.HandleAction(action)
	}
}

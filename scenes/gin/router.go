package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
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
	methods := routeInfo.Methods | scene.HttpMethod(routeInfo.Method)
	if methods == 0 {
		panic("invalid method")
	}
	if methods&scene.HttpMethodGet != 0 {
		r.router.GET(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodHead != 0 {
		r.router.HEAD(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodPost != 0 {
		r.router.POST(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodPut != 0 {
		r.router.PUT(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodPatch != 0 {
		r.router.PATCH(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodDelete != 0 {
		r.router.DELETE(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodOptions != 0 {
		r.router.OPTIONS(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&scene.HttpMethodTrace != 0 {
		panic("trace method not supported method")
	}
}

func (r *AppRouter[T]) HandleActions(actions ...Action[T]) {
	for _, action := range actions {
		r.HandleAction(action)
	}
}

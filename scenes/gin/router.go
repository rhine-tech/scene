package gin

import (
	"github.com/gin-gonic/gin"
)

type Action[T any] interface {
	Request[T]
	HttpRoute
	Middleware() gin.HandlersChain
}

type BaseAction struct {
}

func (a *BaseAction) Middleware() gin.HandlersChain {
	return nil
}

type AppRouter[T any] struct {
	app         T
	router      gin.IRouter
	middlewares gin.HandlersChain
}

func NewAppRouter[T any](app T, router gin.IRouter, middlewares gin.HandlersChain) *AppRouter[T] {
	return &AppRouter[T]{app: app, router: router}
}

func (r *AppRouter[T]) Router() gin.IRouter {
	return r.router
}

func (r *AppRouter[T]) HandleAction(action Action[T]) {
	routeInfo := action.GetRoute()
	chain := MiddlewareChain(append(r.middlewares, action.Middleware()...)...)
	methods := routeInfo.Methods | HttpMethod(routeInfo.Method)
	if methods == 0 {
		panic("invalid method")
	}
	if methods&HttpMethodGet != 0 {
		r.router.GET(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodHead != 0 {
		r.router.HEAD(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodPost != 0 {
		r.router.POST(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodPut != 0 {
		r.router.PUT(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodPatch != 0 {
		r.router.PATCH(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodDelete != 0 {
		r.router.DELETE(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodOptions != 0 {
		r.router.OPTIONS(routeInfo.Path, chain(Handle(r.app, action))...)
	}
	if methods&HttpMethodTrace != 0 {
		panic("trace method not supported method")
	}
}

func (r *AppRouter[T]) HandleActions(actions ...Action[T]) {
	for _, action := range actions {
		r.HandleAction(action)
	}
}

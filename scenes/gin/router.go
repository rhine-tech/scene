package gin

import "github.com/gin-gonic/gin"

// Action describes one HTTP endpoint.
//
// Binding and middleware are optional. An action can additionally implement
// BindingProvider and MiddlewareProvider.
type Action[T any] interface {
	Process(ctx *Context[T]) (data any, err error)
	HttpRoute
}

// MiddlewareProvider supplies middleware that applies only to one action.
type MiddlewareProvider interface {
	Middleware() gin.HandlersChain
}

// AppRouter registers actions for one application context.
type AppRouter[T any] struct {
	app         T
	router      gin.IRouter
	middlewares gin.HandlersChain
}

// NewAppRouter creates a router for app and takes an immutable copy of the
// application middleware chain.
func NewAppRouter[T any](app T, router gin.IRouter, middlewares gin.HandlersChain) *AppRouter[T] {
	return &AppRouter[T]{
		app:         app,
		router:      router,
		middlewares: append(gin.HandlersChain(nil), middlewares...),
	}
}

// Router returns the underlying Gin router.
func (r *AppRouter[T]) Router() gin.IRouter {
	return r.router
}

// HandleAction registers every method declared by action's method bitmap.
func (r *AppRouter[T]) HandleAction(action Action[T]) {
	routeInfo := action.GetRoute()
	if routeInfo.Methods == 0 {
		panic("scene-gin: action must declare at least one HTTP method")
	}
	if unsupported := routeInfo.Methods &^ httpMethodAll; unsupported != 0 {
		panic("scene-gin: action declares an unsupported HTTP method")
	}

	actionMiddlewares := gin.HandlersChain(nil)
	if provider, ok := action.(MiddlewareProvider); ok {
		actionMiddlewares = provider.Middleware()
	}
	handlers := make(
		gin.HandlersChain,
		0,
		len(r.middlewares)+len(actionMiddlewares)+1,
	)
	handlers = append(handlers, r.middlewares...)
	handlers = append(handlers, actionMiddlewares...)
	handlers = append(handlers, Handle(r.app, action))

	for _, method := range httpMethods {
		if routeInfo.Methods&method.bitmap != 0 {
			r.router.Handle(method.name, routeInfo.Path, handlers...)
		}
	}
}

// HandleActions registers actions in order.
func (r *AppRouter[T]) HandleActions(actions ...Action[T]) {
	for _, action := range actions {
		r.HandleAction(action)
	}
}

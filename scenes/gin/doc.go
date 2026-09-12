// Package gin integrates Gin HTTP routing with Scene applications, dependency
// injection, request binding, middleware, and the common response envelope.
//
// # Routers
//
// A Scene owns one HTTP listener and explicitly configured, independent Gin
// routing trees. ModuleLoader still owns every application and its dependency
// injection and module lifecycle:
//
//	sgin.NewFactory(":8080",
//		sgin.DefaultRouter("/api", sgin.WithRecovery()),
//		sgin.Router("admin", "/spa", sgin.WithRecovery()),
//		sgin.Router("web", "/", sgin.WithRecovery()),
//	)
//
// Applications without RouterSelector belong to DefaultRouter.
// AppRoutes.Router and RouterSelector.RouterName select another configured
// router. Unknown names, duplicate names, and duplicate normalized prefixes
// are configuration errors. Each router has its own middleware and NoRoute.
//
// Requests select the longest prefix at a path-segment boundary: /api matches
// /api and /api/users, but not /api2. A router's response is final, including
// 404 and 405; the root router is not a fallback for another router's errors.
// Dispatch does not clean or rewrite request paths. Application Create/Destroy
// follow module declaration order, independently of prefix matching order.
//
// Replace NewFactory(addr, prefix, options...) with
// NewFactory(addr, DefaultRouter(prefix, options...)) when migrating a single
// routing tree. Router options replace the former Scene-wide Gin options.
//
// # Applications
//
// AppRoutes is the usual entry point. It owns an application context, injects
// that context when the application is created, and mounts every action below
// BasePath:
//
//	app := &sgin.AppRoutes[appContext]{
//		AppName:  moduleName.ImplNameNoVer("GinApplication"),
//		BasePath: "users",
//		Context:  appContext{},
//		Actions: []sgin.Action[*appContext]{
//			new(getUserAction),
//		},
//	}
//
// An action only needs route metadata and Process:
//
//	type healthAction struct{}
//
//	func (*healthAction) GetRoute() sgin.HttpRouteInfo {
//		return sgin.HttpRouteInfo{
//			Methods: sgin.HttpMethodGet,
//			Path:    "/health",
//		}
//	}
//
//	func (*healthAction) Process(
//		ctx *sgin.Context[*appContext],
//	) (any, error) {
//		return map[string]string{"status": "ok"}, nil
//	}
//
// # Direct Gin registration
//
// Implement GinApplication directly when an endpoint should use Gin's native
// handlers without Action, binding, or Scene's response envelope. Create
// receives both the selected Router's *gin.Engine and a *gin.RouterGroup
// already scoped by that Router's prefix and the application's Prefix:
//
//	type rawGinApplication struct{}
//
//	func (*rawGinApplication) Name() scene.ImplName {
//		return moduleName.ImplNameNoVer("RawGinApplication")
//	}
//
//	func (*rawGinApplication) Prefix() string {
//		return "raw"
//	}
//
//	func (*rawGinApplication) Create(
//		engine *gin.Engine,
//		router gin.IRouter,
//	) error {
//		router.GET("/health", func(ctx *gin.Context) {
//			ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
//		})
//
//		// Native engine registration can bypass the application prefix,
//		// but the path must still belong to this Router (here, /api).
//		engine.GET("/api/ready", func(ctx *gin.Context) {
//			ctx.Status(http.StatusNoContent)
//		})
//		return nil
//	}
//
//	func (*rawGinApplication) Destroy() error {
//		return nil
//	}
//
// # Binding
//
// Binding is optional. Embed RequestJson, RequestQuery, RequestURI, or another
// request helper for one source. Implement BindingProvider when an action
// needs multiple sources:
//
//	func (*updateUserAction) Bindings() []sgin.Binding {
//		return []sgin.Binding{
//			sgin.BindURI,
//			sgin.BindJSON,
//		}
//	}
//
// Bindings run in declaration order. The explicit URI, query, JSON, and form
// bindings defer validation until every binding has populated the action, so
// binding:"required" works across multiple sources.
//
// # Middleware
//
// AppRoutes.Middlewares apply to every application action. An individual
// action can implement MiddlewareProvider to append route-specific
// middleware. Router middleware runs first, followed by application
// middleware, action middleware, and Process.
//
// # Context and responses
//
// Context embeds *gin.Context, exposes the injected application context as
// App, and implements context.Context by delegating cancellation and values to
// the request. Process results use Scene's common response envelope. An action
// that writes a streaming or otherwise custom response can return
// ErrAlreadyDone to prevent the default renderer from writing another body.
package gin

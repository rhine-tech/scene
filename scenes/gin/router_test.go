package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type middlewareTestApp struct {
	order []string
}

type methodBitmapTestApp struct {
	methods []string
}

type methodBitmapTestAction struct{}

func (a *methodBitmapTestAction) GetRoute() HttpRouteInfo {
	return HttpRouteInfo{
		Methods: HttpMethodGet | HttpMethodPost,
		Path:    "/methods",
	}
}

func (a *methodBitmapTestAction) Process(ctx *Context[*methodBitmapTestApp]) (any, error) {
	ctx.App.methods = append(ctx.App.methods, ctx.Request.Method)
	return nil, nil
}

type middlewareTestAction struct {
	middleware gin.HandlerFunc
}

func (a *middlewareTestAction) GetRoute() HttpRouteInfo {
	return HttpRouteInfo{
		Methods: HttpMethodGet,
		Path:    "/middleware",
	}
}

func (a *middlewareTestAction) Middleware() gin.HandlersChain {
	return gin.HandlersChain{a.middleware}
}

func (a *middlewareTestAction) Process(ctx *Context[*middlewareTestApp]) (any, error) {
	ctx.App.order = append(ctx.App.order, "handler")
	return nil, nil
}

func TestAppRouterAppliesApplicationAndActionMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	app := &middlewareTestApp{}

	appMiddleware := func(ctx *gin.Context) {
		app.order = append(app.order, "application")
		ctx.Next()
	}
	actionMiddleware := func(ctx *gin.Context) {
		app.order = append(app.order, "action")
		ctx.Next()
	}

	router := NewAppRouter(
		app,
		engine.Group("/api"),
		gin.HandlersChain{appMiddleware},
	)
	router.HandleAction(&middlewareTestAction{middleware: actionMiddleware})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/middleware", nil)
	engine.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []string{"application", "action", "handler"}, app.order)
}

func TestAppRouterRegistersMethodBitmap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	app := &methodBitmapTestApp{}
	router := NewAppRouter(app, engine.Group("/api"), nil)
	router.HandleAction(new(methodBitmapTestAction))

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, "/api/methods", nil)
		engine.ServeHTTP(recorder, request)
		require.Equal(t, http.StatusOK, recorder.Code)
	}

	require.Equal(t, []string{http.MethodGet, http.MethodPost}, app.methods)
}

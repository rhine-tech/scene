package gin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/spa/delivery"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

type routerTestApplication struct {
	name   string
	events *[]string
}

func (a *routerTestApplication) Name() scene.ImplName {
	return scene.NewModuleImplNameNoVer("test", a.name)
}

func (a *routerTestApplication) Prefix() string { return "/status" }

func (a *routerTestApplication) Setup() error {
	*a.events = append(*a.events, "setup "+a.name)
	return nil
}

func (a *routerTestApplication) TearDown() error {
	*a.events = append(*a.events, "teardown "+a.name)
	return nil
}

func (a *routerTestApplication) Create(engine *gin.Engine, router gin.IRouter) error {
	*a.events = append(*a.events, "create "+a.name)
	router.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, a.name) })
	// Even native engine registration cannot receive another Router's requests.
	engine.GET("/outside", func(c *gin.Context) { c.String(http.StatusOK, a.name) })
	return nil
}

func (a *routerTestApplication) Destroy() error {
	*a.events = append(*a.events, "destroy "+a.name)
	return nil
}

type selectedRouterTestApplication struct {
	*routerTestApplication
	routerName string
}

func (a *selectedRouterTestApplication) RouterName() string { return a.routerName }

type routerTestModule struct {
	scene.ModuleFactory
	apps []scene.Application
}

func (m routerTestModule) Apps() []scene.Application { return m.apps }

func TestRouterIsolationAndApplicationLifecycle(t *testing.T) {
	var events []string
	files := func(index string) fstest.MapFS {
		return fstest.MapFS{
			"index.html": {Data: []byte(index)},
			"app.js":     {Data: []byte("app script")},
		}
	}
	loader := scene.NewModuleLoaderIn(registry.NewScope(), scene.ModuleFactoryArray{
		routerTestModule{apps: []scene.Application{
			&selectedRouterTestApplication{&routerTestApplication{"admin", &events}, "admin"},
			&routerTestApplication{"api", &events},
			delivery.NewGinSPA(files("main SPA"), "web", "."),
			delivery.NewGinSPA(files("admin SPA"), "admin", "."),
		}},
	})
	require.NoError(t, loader.Init())
	t.Cleanup(func() { require.NoError(t, loader.TearDown()) })
	require.NoError(t, loader.Setup())
	markRouter := func(name string) GinOption {
		return func(_ *registry.Scope, engine *gin.Engine) error {
			engine.HandleMethodNotAllowed = true
			engine.Use(func(c *gin.Context) { c.Header("X-Router", name) })
			return nil
		}
	}
	definition := NewFactory("127.0.0.1:0",
		Router("web", "/", markRouter("web")),
		DefaultRouter("api/", markRouter("api")),
		Router("admin", "/spa/", markRouter("admin")),
	)
	built, err := definition.Build(loader)
	require.NoError(t, err)
	container := built.(*ginContainer)
	t.Cleanup(container.cancel)
	require.NoError(t, container.startApps())

	for _, test := range []struct {
		method string
		path   string
		status int
		body   string
		router string
	}{
		{http.MethodGet, "/api/status/ping", 200, "api", "api"},
		{http.MethodGet, "/api", 404, "404 page not found", "api"},
		{http.MethodGet, "/api/unknown", 404, "404 page not found", "api"},
		{http.MethodPost, "/api/status/ping", 405, "405 method not allowed", "api"},
		{http.MethodGet, "/spa/status/ping", 200, "admin", "admin"},
		{http.MethodGet, "/spa", 200, "admin SPA", "admin"},
		{http.MethodGet, "/spa/", 200, "admin SPA", "admin"},
		{http.MethodGet, "/spa/projects/123", 200, "admin SPA", "admin"},
		{http.MethodHead, "/spa/projects/123", 200, "", "admin"},
		{http.MethodGet, "/spa/app.js", 200, "app script", "admin"},
		{http.MethodGet, "/spa/missing.js", 404, "404 page not found", "admin"},
		{http.MethodPost, "/spa/projects/123", 404, "404 page not found", "admin"},
		{http.MethodGet, "/", 200, "main SPA", "web"},
		{http.MethodGet, "/projects/123", 200, "main SPA", "web"},
		{http.MethodGet, "/api2/status/ping", 200, "main SPA", "web"},
		{http.MethodGet, "/outside", 200, "main SPA", "web"},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			container.handler.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
			require.Equal(t, test.body, response.Body.String())
			require.Equal(t, test.router, response.Header().Get("X-Router"))
			require.Equal(t, test.path, request.URL.Path)
		})
	}
	require.NoError(t, container.stopApps())
	require.NoError(t, loader.TearDown())
	require.Equal(t, []string{
		"setup admin", "setup api", "create admin", "create api",
		"destroy admin", "destroy api", "teardown api", "teardown admin",
	}, events)
}

func TestRouterConfiguration(t *testing.T) {
	for _, test := range []struct {
		name    string
		routers []RouterDefinition
		apps    []GinApplication
		err     string
	}{
		{"duplicate name", []RouterDefinition{Router("web", "/"), Router("web", "/spa")}, nil, "duplicate router name"},
		{"duplicate prefix", []RouterDefinition{DefaultRouter("/api"), Router("other", "api/")}, nil, "same prefix"},
		{"unknown router", []RouterDefinition{DefaultRouter("/api")}, []GinApplication{
			&selectedRouterTestApplication{&routerTestApplication{name: "app"}, "missing"},
		}, `unconfigured router "missing"`},
		{"missing default", []RouterDefinition{Router("web", "/")}, []GinApplication{
			&routerTestApplication{name: "app"},
		}, `unconfigured router ""`},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := (Factory{Routers: test.routers}).Build(registry.NewScope(), test.apps)
			require.ErrorContains(t, err, test.err)
		})
	}
}

func TestHTTPMiddlewareRouting(t *testing.T) {
	var events []string
	scope := registry.NewScope()
	registry.SetIn[http.Handler](scope, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		events = append(events, "site")
		if r.URL.Path != "/api/users" {
			http.Error(w, "site not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("site"))
	}))
	trace := func(name string) HTTPMiddleware {
		return func(_ *registry.Scope, next http.Handler) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				events = append(events, name+" before")
				next.ServeHTTP(w, r)
				events = append(events, name+" after")
			}), nil
		}
	}
	router := func(name, route string) GinOption {
		return func(_ *registry.Scope, engine *gin.Engine) error {
			engine.Use(func(c *gin.Context) { events = append(events, name) })
			engine.GET(route, func(c *gin.Context) { c.String(http.StatusOK, name) })
			return nil
		}
	}
	built, err := (Factory{
		Routers: []RouterDefinition{
			Router("web", "/", router("web", "/")),
			DefaultRouter("/api", router("api", "/api/users")),
		},
		HTTPMiddlewares: []HTTPMiddleware{
			trace("outer"),
			func(scope *registry.Scope, next http.Handler) (http.Handler, error) {
				site := registry.UseIn[http.Handler](scope, nil)
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Host == "site.example.com" {
						site.ServeHTTP(w, r)
						return
					}
					next.ServeHTTP(w, r)
				}), nil
			},
			trace("inner"),
		},
	}).Build(scope, nil)
	require.NoError(t, err)
	container := built.(*ginContainer)
	t.Cleanup(container.cancel)

	for _, test := range []struct {
		name   string
		host   string
		path   string
		status int
		body   string
		events []string
	}{
		{"site before API", "site.example.com", "/api/users", 200, "site", []string{"outer before", "site", "outer after"}},
		{"site 404 is final", "site.example.com", "/api/missing", 404, "site not found\n", []string{"outer before", "site", "outer after"}},
		{"API", "app.example.com", "/api/users", 200, "api", []string{"outer before", "inner before", "api", "inner after", "outer after"}},
		{"API 404 is final", "app.example.com", "/api/missing", 404, "404 page not found", []string{"outer before", "inner before", "api", "inner after", "outer after"}},
		{"root", "app.example.com", "/", 200, "web", []string{"outer before", "inner before", "web", "inner after", "outer after"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			events = nil
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Host = test.host
			response := httptest.NewRecorder()
			container.handler.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
			require.Equal(t, test.body, response.Body.String())
			require.Equal(t, test.events, events)
		})
	}
}

func TestHTTPMiddlewareBuildError(t *testing.T) {
	want := errors.New("gateway configuration failed")
	built, err := (Factory{
		HTTPMiddlewares: []HTTPMiddleware{
			func(*registry.Scope, http.Handler) (http.Handler, error) {
				return nil, want
			},
		},
	}).Build(registry.NewScope(), nil)
	require.ErrorIs(t, err, want)
	require.Nil(t, built)
}

type routerBenchmarkWriter struct {
	header http.Header
}

func (w *routerBenchmarkWriter) Header() http.Header         { return w.header }
func (w *routerBenchmarkWriter) WriteHeader(int)             {}
func (w *routerBenchmarkWriter) Write(p []byte) (int, error) { return len(p), nil }

func BenchmarkRouterDispatch(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	api := gin.New()
	api.GET("/api/projects/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	mux := prefixMux{
		{prefix: "/api", engine: api},
		{prefix: "/spa", engine: gin.New()},
		{prefix: "/", engine: gin.New()},
	}
	for _, variant := range []struct {
		name    string
		handler http.Handler
	}{
		{"Direct", api},
		{"Router", mux},
	} {
		b.Run(variant.name, func(b *testing.B) {
			request := httptest.NewRequest(http.MethodGet, "/api/projects/123", nil)
			writer := &routerBenchmarkWriter{header: make(http.Header)}
			variant.handler.ServeHTTP(writer, request)
			b.ReportAllocs()
			for b.Loop() {
				variant.handler.ServeHTTP(writer, request)
			}
		})
	}
}

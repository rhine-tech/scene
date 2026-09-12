package delivery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func spaTestFiles() fstest.MapFS {
	return fstest.MapFS{
		"dist/index.html":             {Data: []byte("<!doctype html><title>App</title>\n")},
		"dist/config.production.json": {Data: []byte("{\"api\":\"/api\"}\n")},
	}
}

func TestGinSPAMount(t *testing.T) {
	files := spaTestFiles()
	for _, prefix := range []string{"/", "/ui"} {
		t.Run(prefix, func(t *testing.T) {
			app := NewGinSPA(files, "web", "dist")
			engine := gin.New()
			require.NoError(t, app.Create(engine, engine.Group(prefix)))
			base := strings.TrimRight(prefix, "/")
			for _, resource := range []struct {
				path string
				file string
			}{
				{prefix, "index.html"},
				{base + "/projects/123", "index.html"},
				{base + "/config.production.json", "config.production.json"},
			} {
				content, err := files.ReadFile("dist/" + resource.file)
				require.NoError(t, err)
				for _, method := range []string{http.MethodGet, http.MethodHead} {
					response := httptest.NewRecorder()
					engine.ServeHTTP(response, httptest.NewRequest(method, resource.path, nil))
					require.Equal(t, http.StatusOK, response.Code, "%s %s", method, resource.path)
					if method == http.MethodHead {
						require.Empty(t, response.Body.String())
					} else {
						require.Equal(t, string(content), response.Body.String())
					}
				}
			}
		})
	}
}

func TestGinSPACacheValidation(t *testing.T) {
	files := spaTestFiles()
	app := NewGinSPA(files, "web", "dist")
	engine := gin.New()
	require.NoError(t, app.Create(engine, engine.Group("/ui")))

	for _, test := range []struct {
		path         string
		file         string
		cacheControl string
	}{
		{"/ui/", "index.html", "no-cache, must-revalidate"},
		{"/ui/projects/123", "index.html", "no-cache, must-revalidate"},
		{"/ui/config.production.json", "config.production.json", "public, max-age=3600"},
	} {
		t.Run(test.path, func(t *testing.T) {
			content, err := files.ReadFile("dist/" + test.file)
			require.NoError(t, err)
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			require.Equal(t, http.StatusOK, response.Code)
			require.Equal(t, string(content), response.Body.String())
			require.Equal(t, test.cacheControl, response.Header().Get("Cache-Control"))
			etag := response.Header().Get("ETag")
			require.NotEmpty(t, etag)

			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Header.Set("If-None-Match", etag)
			response = httptest.NewRecorder()
			engine.ServeHTTP(response, request)
			require.Equal(t, http.StatusNotModified, response.Code)
			require.Empty(t, response.Body.String())
			require.Equal(t, etag, response.Header().Get("ETag"))
		})
	}
}

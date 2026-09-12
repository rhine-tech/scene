package gin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RouterDefinition declares an independent Gin routing tree and its URL prefix.
// Name is selected by RouterSelector. Use DefaultRouter for the default tree.
type RouterDefinition struct {
	Name    string
	Prefix  string
	Options []GinOption
}

// Router declares a routing tree. Its options apply only to that tree,
// including its NoRoute and NoMethod handlers.
func Router(name, prefix string, options ...GinOption) RouterDefinition {
	return RouterDefinition{Name: name, Prefix: prefix, Options: options}
}

// DefaultRouter declares the routing tree used by applications that do not
// select a named Router.
func DefaultRouter(prefix string, options ...GinOption) RouterDefinition {
	return Router("", prefix, options...)
}

type ginRouter struct {
	name   string
	prefix string
	engine *gin.Engine
}

// prefixMux is sorted by descending prefix length during Factory.Build.
// It selects one router without rewriting the request or retrying on a 404.
type prefixMux []*ginRouter

func (mux prefixMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	for _, router := range mux {
		prefix := router.prefix
		if prefix == "/" || requestPath == prefix ||
			(len(requestPath) > len(prefix) && requestPath[len(prefix)] == '/' && strings.HasPrefix(requestPath, prefix)) {
			router.engine.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

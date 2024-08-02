package gin

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// WithGzip add gzip support for gin engine,
// default level is -1
func WithGzip(level int) GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(gzip.Gzip(level))
		return nil
	}
}

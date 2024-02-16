package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
)

type Context[T GinApplication] struct {
	*gin.Context
	App T
}

func (g *Context[T]) Get(key string) (value any, exists bool) {
	return g.Context.Get(key)
}

func (g *Context[T]) Set(key string, value any) {
	g.Context.Set(key, value)
}

func GetContext(c *gin.Context) scene.Context {
	return &Context[GinApplication]{Context: c, App: GinApplication(nil)}
}

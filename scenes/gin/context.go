package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
)

type ginCtx struct {
	c *gin.Context
}

func (g *ginCtx) Get(key string) (value any, exists bool) {
	return g.c.Get(key)
}

func (g *ginCtx) Set(key string, value any) {
	g.c.Set(key, value)
}

func GetContext(c *gin.Context) scene.Context {
	return &ginCtx{c: c}
}

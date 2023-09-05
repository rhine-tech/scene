package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
)

type ginAppFactory struct {
	engine *gin.Engine
}

func NewAppFactory(engine *gin.Engine) scene.ApplicationFactory[GinApplication] {
	return &ginAppFactory{
		engine: engine,
	}
}

func (g *ginAppFactory) Name() string {
	return "scene.factory.gin"
}

func (g *ginAppFactory) Create(t GinApplication) error {
	return t.(GinApplication).Create(g.engine, g.engine.Group(t.Prefix()))
}

func (g *ginAppFactory) Destroy(t GinApplication) error {
	return t.(GinApplication).Destroy()
}

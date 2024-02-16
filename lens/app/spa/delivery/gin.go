package delivery

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"io/fs"
	"mime"
	"path/filepath"
)

type GinSPA struct {
	sgin.CommonApp
	fsEmbed fs.ReadFileFS
}

func NewGinSPA(fsEmbed *embed.FS, fsPrefix string) *GinSPA {
	fsys, err := fs.Sub(fsEmbed, fsPrefix)
	if err != nil {
		panic(err)
	}
	return &GinSPA{
		fsEmbed: fsys.(fs.ReadFileFS),
	}
}

func (g *GinSPA) Name() scene.ImplName {
	return scene.NewAppImplNameNoVer("spa", "gin")
}

func (g *GinSPA) Prefix() string {
	return "/"
}

func (g *GinSPA) Create(engine *gin.Engine, router gin.IRouter) error {
	engine.NoRoute(g.handleSpa)
	return nil
}

func (g *GinSPA) Destroy() error {
	return nil
}

func (g *GinSPA) handleSpa(c *gin.Context) {
	path := c.Request.URL.Path
	if len(path) > 0 {
		path = path[1:]
	}
	if data, err := g.fsEmbed.ReadFile(path); err == nil {
		c.Data(200, mime.TypeByExtension(filepath.Ext(path)), data)
		return
	}
	if data, err := g.fsEmbed.ReadFile("index.html"); err != nil {
		c.String(404, "404 Not Found")
	} else {
		c.Data(200, mime.TypeByExtension(".html"), data)
	}
}

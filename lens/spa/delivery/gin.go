package delivery

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"io/fs"
	"net/http"
	"strings"
)

type GinSPA struct {
	fsEmbed   fs.ReadFileFS
	handler   http.Handler
	urlPrefix string
}

func NewGinSPA(fsEmbed *embed.FS, urlPrefix, fsPrefix string) *GinSPA {
	fsys, err := fs.Sub(fsEmbed, fsPrefix)
	if err != nil {
		panic(err)
	}
	urlPrefix = "/" + strings.TrimLeft(urlPrefix, "/")
	return &GinSPA{
		fsEmbed:   fsys.(fs.ReadFileFS),
		handler:   http.StripPrefix(urlPrefix, http.FileServer(http.FS(fsys))),
		urlPrefix: urlPrefix,
	}
}

func (g *GinSPA) Name() scene.ImplName {
	return scene.NewModuleImplNameNoVer("spa", "gin")
}

func (g *GinSPA) Prefix() string {
	return g.urlPrefix
}

// It uses a catch-all route to handle all requests under the specified prefix.
func (g *GinSPA) Create(engine *gin.Engine, router gin.IRouter) error {
	engine.Any(g.urlPrefix+"/*any", g.handleSpa)
	return nil
}

func (g *GinSPA) Destroy() error {
	return nil
}

// handleSpa is the core logic for serving the SPA.
func (g *GinSPA) handleSpa(c *gin.Context) {
	requestPath := strings.TrimPrefix(c.Request.URL.Path, g.urlPrefix)
	requestPath = strings.TrimLeft(requestPath, "/")
	// 1. CHECK IF FILE EXISTS by trying to open it.
	_, err := g.fsEmbed.Open(requestPath)
	if err != nil {
		c.Request.URL.Path = g.urlPrefix + "/"
	}
	g.handler.ServeHTTP(c.Writer, c.Request)
}

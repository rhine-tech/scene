package delivery

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type GinSPA struct {
	handler    http.Handler
	routerName string
	urlPrefix  string
	files      map[string]string
}

// NewGinSPA serves fsPrefix inside files using the named Gin Router's NoRoute
// handler. Its URL prefix is supplied by that Router when Create is called.
func NewGinSPA(files fs.FS, routerName, fsPrefix string) *GinSPA {
	fsys, err := fs.Sub(files, fsPrefix)
	if err != nil {
		panic(err)
	}
	etags, err := buildSPAETags(fsys)
	if err != nil {
		panic(err)
	}
	return &GinSPA{
		handler:    http.FileServer(http.FS(fsys)),
		routerName: routerName,
		files:      etags,
	}
}

func (g *GinSPA) Name() scene.ImplName {
	return scene.NewModuleImplNameNoVer("spa", "gin")
}

func (g *GinSPA) Prefix() string {
	return "/"
}

func (g *GinSPA) RouterName() string {
	return g.routerName
}

// Create installs SPA fallback in its own Router. The Scene supplies a
// *gin.RouterGroup containing the Router's configured URL prefix.
func (g *GinSPA) Create(engine *gin.Engine, router gin.IRouter) error {
	g.urlPrefix = strings.TrimSuffix(router.(*gin.RouterGroup).BasePath(), "/")
	engine.NoRoute(g.handleSpa)
	return nil
}

func (g *GinSPA) Destroy() error {
	return nil
}

// handleSpa is the core logic for serving the SPA.
func (g *GinSPA) handleSpa(c *gin.Context) {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		c.Status(http.StatusNotFound)
		return
	}
	requestPath := strings.TrimPrefix(c.Request.URL.Path, g.urlPrefix)
	requestPath = normalizeRequestPath(requestPath)

	targetPath := requestPath
	if targetPath == "" {
		targetPath = "index.html"
	}

	etag, found := g.files[targetPath]
	if !found {
		// Only route-style paths (without file extension) should fall back to index.html.
		// Missing static assets should return 404 to avoid caching wrong content.
		if !shouldFallbackToIndex(requestPath) {
			c.Header("Cache-Control", "no-cache")
			c.Status(http.StatusNotFound)
			return
		}
		targetPath = "index.html"
		etag = g.files[targetPath]
	}

	if targetPath == "index.html" {
		c.Header("Cache-Control", "no-cache, must-revalidate")
	} else {
		c.Header("Cache-Control", "public, max-age=3600")
	}
	c.Header("ETag", etag)

	originalPath, originalRawPath := c.Request.URL.Path, c.Request.URL.RawPath
	if targetPath == "index.html" {
		c.Request.URL.Path = "/"
	} else {
		c.Request.URL.Path = "/" + targetPath
	}
	c.Request.URL.RawPath = ""
	g.handler.ServeHTTP(c.Writer, c.Request)
	c.Request.URL.Path, c.Request.URL.RawPath = originalPath, originalRawPath
}

func buildSPAETags(fsys fs.FS) (map[string]string, error) {
	files := make(map[string]string)
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		file, err := fsys.Open(p)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		sum := sha256.New()
		if _, err = io.Copy(sum, file); err != nil {
			return err
		}
		files[p] = `"` + hex.EncodeToString(sum.Sum(nil)) + `"`
		return nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := files["index.html"]; !ok {
		return nil, fs.ErrNotExist
	}
	return files, nil
}

func normalizeRequestPath(raw string) string {
	cleaned := path.Clean("/" + strings.TrimSpace(raw))
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func shouldFallbackToIndex(requestPath string) bool {
	if requestPath == "" {
		return true
	}
	base := path.Base(requestPath)
	return path.Ext(base) == ""
}

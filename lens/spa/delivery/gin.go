package delivery

import (
	"crypto/sha256"
	"embed"
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
	handler   http.Handler
	urlPrefix string
	files     map[string]spaFileMeta
}

type spaFileMeta struct {
	etag      string
	immutable bool
}

func NewGinSPA(fsEmbed *embed.FS, urlPrefix, fsPrefix string) *GinSPA {
	fsys, err := fs.Sub(fsEmbed, fsPrefix)
	if err != nil {
		panic(err)
	}
	metas, err := buildSPAFilesMeta(fsys)
	if err != nil {
		panic(err)
	}
	urlPrefix = "/" + strings.TrimLeft(urlPrefix, "/")
	return &GinSPA{
		handler:   http.StripPrefix(urlPrefix, http.FileServer(http.FS(fsys))),
		urlPrefix: urlPrefix,
		files:     metas,
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
	engine.GET(g.urlPrefix, g.handleSpa)
	engine.HEAD(g.urlPrefix, g.handleSpa)
	engine.GET(g.urlPrefix+"/*any", g.handleSpa)
	engine.HEAD(g.urlPrefix+"/*any", g.handleSpa)
	return nil
}

func (g *GinSPA) Destroy() error {
	return nil
}

// handleSpa is the core logic for serving the SPA.
func (g *GinSPA) handleSpa(c *gin.Context) {
	requestPath := strings.TrimPrefix(c.Request.URL.Path, g.urlPrefix)
	requestPath = normalizeRequestPath(requestPath)

	targetPath := requestPath
	if targetPath == "" {
		targetPath = "index.html"
	}

	meta, found := g.files[targetPath]
	if !found {
		// Only route-style paths (without file extension) should fall back to index.html.
		// Missing static assets should return 404 to avoid caching wrong content.
		if !shouldFallbackToIndex(requestPath) {
			c.Header("Cache-Control", "no-cache")
			c.Status(http.StatusNotFound)
			return
		}
		targetPath = "index.html"
		meta = g.files[targetPath]
	}

	if targetPath == "index.html" {
		c.Header("Cache-Control", "no-cache, must-revalidate")
	} else if meta.immutable {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "public, max-age=3600")
	}
	c.Header("ETag", meta.etag)

	if etagMatch(c.GetHeader("If-None-Match"), meta.etag) {
		c.Status(http.StatusNotModified)
		return
	}

	if targetPath == "index.html" {
		c.Request.URL.Path = g.urlPrefix + "/"
	} else {
		c.Request.URL.Path = g.urlPrefix + "/" + targetPath
	}
	g.handler.ServeHTTP(c.Writer, c.Request)
}

func buildSPAFilesMeta(fsys fs.FS) (map[string]spaFileMeta, error) {
	files := make(map[string]spaFileMeta)
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
		files[p] = spaFileMeta{
			etag:      `"` + hex.EncodeToString(sum.Sum(nil)) + `"`,
			immutable: isImmutableAsset(p),
		}
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

func isImmutableAsset(p string) bool {
	base := path.Base(p)
	dot := strings.LastIndexByte(base, '.')
	if dot <= 0 {
		return false
	}
	stem := base[:dot]
	sep := strings.LastIndexAny(stem, ".-")
	if sep <= 0 || sep == len(stem)-1 {
		return false
	}
	token := stem[sep+1:]
	if len(token) < 8 {
		return false
	}
	return isAssetHashToken(token)
}

func isAssetHashToken(s string) bool {
	for _, r := range s {
		if !(r >= '0' && r <= '9' ||
			r >= 'a' && r <= 'z' ||
			r >= 'A' && r <= 'Z' ||
			r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func etagMatch(raw, target string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	if raw == "*" {
		return true
	}
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, "W/")
		if strings.EqualFold(part, target) {
			return true
		}
	}
	return false
}

func shouldFallbackToIndex(requestPath string) bool {
	if requestPath == "" {
		return true
	}
	base := path.Base(requestPath)
	return path.Ext(base) == ""
}

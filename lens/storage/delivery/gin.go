package delivery

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	permMdw "github.com/rhine-tech/scene/lens/permission/middleware"
	"github.com/rhine-tech/scene/lens/storage"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

const (
	dataRoutePath        = "/data/:provider/*fileid"
	urlRoutePath         = "/url/:provider/*fileid"
	urlModeProxy         = "proxy"
	urlModeDirect        = "direct"
	uploadCleanupTimeout = 5 * time.Second
)

type appContext struct {
	srv storage.IStorageService `aperture:""`
}

func GinApp() sgin.GinApplication {
	return &sgin.AppRoutes[appContext]{
		AppName:  storage.Lens.ImplNameNoVer("GinApplication"),
		BasePath: storage.Lens.String(),
		Actions: []sgin.Action[*appContext]{
			new(getDataRequest),
			new(putDataRequest),
			new(deleteDataRequest),
			new(getURLRequest),
			new(listMetaRequest),
			new(listProviderRequest),
		},
		Context: appContext{
			srv: nil,
		},
	}
}

type getDataRequest struct {
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
}

func (l *getDataRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Methods: sgin.HttpMethodGet | sgin.HttpMethodHead | sgin.HttpMethodOptions,
		Path:    dataRoutePath,
	}
}

func (l *getDataRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileDownload),
	}
}

func (l *getDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	reader, meta, err := storage.OpenContent(ctx.Request.Context(), ctx.App.srv, storage.NewStorageKey(l.Provider, l.StorageKey))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	contentType := meta.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx.Header("Content-Type", contentType)
	http.ServeContent(ctx.Writer, ctx.Request, meta.OriginalFilename, meta.UpdatedAt, reader)
	return nil, sgin.ErrAlreadyDone
}

type putDataRequest struct {
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid"`
}

func (p *putDataRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Methods: sgin.HttpMethodPut | sgin.HttpMethodPost,
		Path:    dataRoutePath,
	}
}

func (p *putDataRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileUpload),
	}
}

func (p *putDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	requestCtx := ctx.Request.Context()
	identifier := storage.NormalizeIdentifier(p.StorageKey)
	fileName := ctx.Query("filename")
	if fileName == "" {
		fileName = ctx.Request.Header.Get("filename")
	}
	if fileName == "" {
		fileName = identifier
	}
	if fileName == "" {
		fileName = "upload"
	}
	contentType := ctx.Query("content_type")
	if contentType == "" {
		contentType = ctx.ContentType()
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// Construct metadata (can be expanded from headers as needed)
	meta := storage.FileMeta{
		Provider:         p.Provider,
		OriginalFilename: fileName,
		ContentType:      contentType,
		ContentLength:    ctx.Request.ContentLength,
		Finished:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Init multipart session
	_, uploadId, err := ctx.App.srv.InitMultipartStore(requestCtx, p.Provider, identifier, meta)
	if err != nil {
		return nil, err
	}

	// Store single part from body
	err = ctx.App.srv.StoreMultipart(requestCtx, uploadId, 1, ctx.Request.Body)
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(requestCtx), uploadCleanupTimeout)
		defer cancel()
		_ = ctx.App.srv.AbortMultipart(cleanupCtx, uploadId)
		return nil, err
	}

	// Complete the multipart upload
	return ctx.App.srv.CompleteMultipart(requestCtx, uploadId)
}

type deleteDataRequest struct {
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
}

func (d *deleteDataRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Methods: sgin.HttpMethodDelete,
		Path:    dataRoutePath,
	}
}

func (d *deleteDataRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileDelete),
	}
}

func (d *deleteDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	d.StorageKey = strings.TrimPrefix(d.StorageKey, "/")
	storageKey := storage.NewStorageKey(d.Provider, d.StorageKey)
	if err := ctx.App.srv.Delete(ctx.Request.Context(), storageKey); err != nil {
		return nil, err
	}
	return storage.FileMeta{StorageKey: storageKey}, nil
}

type getURLRequest struct {
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
	mode       string
}

func (g *getURLRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Methods: sgin.HttpMethodGet,
		Path:    urlRoutePath,
	}
}

func (g *getURLRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileDownload),
	}
}

func (g *getURLRequest) Bindings() []sgin.Binding {
	return []sgin.Binding{sgin.BindURI, g.bindMode}
}

func (g *getURLRequest) bindMode(ctx *gin.Context, _ any) error {
	g.mode = ctx.DefaultQuery("mode", urlModeProxy)
	if g.mode != urlModeProxy && g.mode != urlModeDirect {
		return fmt.Errorf("unsupported URL mode %q", g.mode)
	}
	return nil
}

func (g *getURLRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	g.StorageKey = strings.TrimPrefix(g.StorageKey, "/")
	storageKey := storage.NewStorageKey(g.Provider, g.StorageKey)
	if err := storage.ValidateStorageKey(storageKey); err != nil {
		return nil, err
	}
	if g.mode == urlModeDirect {
		return ctx.App.srv.GetDirectURL(ctx.Request.Context(), storageKey)
	}

	routePrefix, ok := strings.CutSuffix(ctx.FullPath(), urlRoutePath)
	if !ok {
		return nil, fmt.Errorf("unexpected storage URL route %q", ctx.FullPath())
	}
	return url.JoinPath(routePrefix, "data", storageKey.Provider(), storageKey.FileID())
}

type listMetaRequest struct {
	Provider string `uri:"provider" binding:"required"`
	Offset   int64  `form:"offset,default=0"`
	Limit    int64  `form:"limit,default=20" binding:"required"`
}

func (l *listMetaRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Methods: sgin.HttpMethodGet,
		Path:    "/list/:provider",
	}
}

func (l *listMetaRequest) Bindings() []sgin.Binding {
	return []sgin.Binding{sgin.BindURI, sgin.BindQuery}
}

func (l *listMetaRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileList),
	}
}

func (l *listMetaRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	return ctx.App.srv.ListMeta(ctx.Request.Context(), l.Provider, l.Offset, l.Limit)
}

type listProviderRequest struct{}

func (l *listProviderRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Methods: sgin.HttpMethodGet,
		Path:    "/providers",
	}
}

func (l *listProviderRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileList),
	}
}

func (l *listProviderRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	return ctx.App.srv.ListProviders(), nil
}

package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/lens/permission"
	permMdw "github.com/rhine-tech/scene/lens/permission/middleware"
	"github.com/rhine-tech/scene/lens/storage"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
	"strings"
	"time"
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
			new(getPublicURLRequest),
			new(listMetaRequest),
			new(listProviderRequest),
		},
		Context: appContext{
			srv: nil,
		},
	}
}

type getDataRequest struct {
	sgin.BaseAction
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
}

func (l *getDataRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Method:  http.MethodGet,
		Methods: sgin.HttpMethodGet | sgin.HttpMethodHead | sgin.HttpMethodOptions,
		Path:    "/data/:provider/*fileid",
	}
}

func (l *getDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	reader, meta, err := storage.NewIoInterface(ctx.App.srv, storage.NewStorageKey(l.Provider, l.StorageKey))
	if err != nil {
		return nil, err
	}
	http.ServeContent(ctx.Writer, ctx.Request, meta.OriginalFilename, meta.UpdatedAt, reader)
	return nil, sgin.ErrAlreadyDone
}

type putDataRequest struct {
	sgin.BaseAction
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
}

func (p *putDataRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Method:  http.MethodPut,
		Path:    "/data/:provider/*fileid",
		Methods: sgin.HttpMethodPut | sgin.HttpMethodPost,
	}
}

func (p *putDataRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileUpload),
	}
}

func (p *putDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	p.StorageKey = strings.TrimPrefix(p.StorageKey, "/")
	storageKey := storage.NewStorageKey(p.Provider, p.StorageKey)
	if !permission.HasPermissionInCtx(ctx, storage.PermFileNaming) {
		storageKey = storage.NewStorageKeyWithUUID(p.Provider)
	}
	fileName := ctx.Query("filename")
	if fileName == "" {
		fileName = ctx.Request.Header.Get("filename")
	}
	if fileName == "" {
		fileName = storageKey.FileID()
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
		StorageKey:       storageKey,
		Provider:         p.Provider,
		OriginalFilename: fileName,
		ContentType:      contentType,
		ContentLength:    ctx.Request.ContentLength,
		Finished:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Init multipart session
	storageKey, uploadId, err := ctx.App.srv.InitMultipartStore(p.Provider, p.StorageKey, meta)
	if err != nil {
		return nil, err
	}

	// Store single part from body
	err = ctx.App.srv.StorePartReader(uploadId, 1, ctx.Request.Body)
	if err != nil {
		_ = ctx.App.srv.AbortMultiPartStore(uploadId) // cleanup on failure
		return nil, err
	}

	// Complete the multipart upload
	err = ctx.App.srv.CompleteMultipartStore(uploadId)
	if err != nil {
		return nil, err
	}

	meta, err = ctx.App.srv.Meta(storageKey)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

type deleteDataRequest struct {
	sgin.BaseAction
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
}

func (d *deleteDataRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Method: http.MethodDelete,
		Path:   "/data/:provider/*fileid",
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
	if err := ctx.App.srv.Delete(storageKey); err != nil {
		return nil, err
	}
	return storage.FileMeta{StorageKey: storageKey}, nil
}

type getPublicURLRequest struct {
	sgin.BaseAction
	sgin.RequestURI
	Provider   string `uri:"provider" binding:"required"`
	StorageKey string `uri:"fileid" binding:"required"`
}

func (g *getPublicURLRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Method: http.MethodGet,
		Path:   "/url/:provider/*fileid",
	}
}

func (g *getPublicURLRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileDownload),
	}
}

func (g *getPublicURLRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	g.StorageKey = strings.TrimPrefix(g.StorageKey, "/")
	storageKey := storage.NewStorageKey(g.Provider, g.StorageKey)
	return ctx.App.srv.GetPublicURL(storageKey)
}

type listMetaRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Offset int64 `form:"offset,default=0"`
	Limit  int64 `form:"limit,default=20" binding:"required"`
}

func (l *listMetaRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Method: http.MethodGet,
		Path:   "/list/:provider",
	}
}

func (l *listMetaRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(storage.PermFileList),
	}
}

func (l *listMetaRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	provider := ctx.Param("provider")
	return ctx.App.srv.ListMeta(provider, l.Offset, l.Limit)
}

type listProviderRequest struct {
	sgin.BaseAction
	sgin.RequestNoParam
}

func (l *listProviderRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{
		Method: http.MethodGet,
		Path:   "/providers",
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

package delivery

import (
	"github.com/rhine-tech/scene"
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
	Provider string `uri:"provider" binding:"required"`
	FileID   string `uri:"fileid" binding:"required"`
}

func (l *getDataRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{
		Method:  http.MethodGet,
		Methods: scene.HttpMethodGet | scene.HttpMethodHead | scene.HttpMethodOptions,
		Path:    "/data/:provider/*fileid",
	}
}

func (l *getDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	reader, meta, err := storage.NewIoInterface(ctx.App.srv, storage.NewFileID(l.Provider, l.FileID))
	if err != nil {
		return nil, err
	}
	http.ServeContent(ctx.Writer, ctx.Request, meta.OriginalFilename, meta.UpdatedAt, reader)
	return nil, sgin.ErrAlreadyDone
}

type putDataRequest struct {
	sgin.BaseAction
	sgin.RequestURI
	Provider string `uri:"provider" binding:"required"`
	FileID   string `uri:"fileid" binding:"required"`
}

func (p *putDataRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{
		Method:  http.MethodPut,
		Path:    "/data/:provider/*fileid",
		Methods: scene.HttpMethodPut | scene.HttpMethodPost,
	}
}

func (p *putDataRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	p.FileID = strings.TrimPrefix(p.FileID, "/")
	fileId := storage.NewFileID(p.Provider, p.FileID)
	fileName := ctx.Query("filename")
	if fileName == "" {
		ctx.Request.Header.Get("filename")
	}
	if fileName == "" {
		fileName = p.FileID
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
		FileID:           fileId,
		Provider:         p.Provider,
		OriginalFilename: fileName,
		ContentType:      contentType,
		ContentLength:    ctx.Request.ContentLength,
		Finished:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Init multipart session
	uploadId, err := ctx.App.srv.InitMultipartStore(fileId, meta)
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

	meta, err = ctx.App.srv.Meta(fileId)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

type listMetaRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Offset int64 `form:"offset,default=0"`
	Limit  int64 `form:"limit,default=20" binding:"required"`
}

func (l *listMetaRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{
		Method: http.MethodGet,
		Path:   "/list/:provider",
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

func (l *listProviderRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{
		Method: http.MethodGet,
		Path:   "/providers",
	}
}

func (l *listProviderRequest) Process(ctx *sgin.Context[*appContext]) (data any, err error) {
	return ctx.App.srv.ListProviders(), nil
}

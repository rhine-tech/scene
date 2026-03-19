package arpcimpl

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
)

const (
	ARpcNameStorageIStorageServiceListProviders             = "storage.IStorageService.ListProviders"
	ARpcNameStorageIStorageServiceListMeta                  = "storage.IStorageService.ListMeta"
	ARpcNameStorageIStorageServiceMeta                      = "storage.IStorageService.Meta"
	ARpcNameStorageIStorageServiceDelete                    = "storage.IStorageService.Delete"
	ARpcNameStorageIStorageServiceInitMultipartStore        = "storage.IStorageService.InitMultipartStore"
	ARpcNameStorageIStorageServiceStorePart                 = "storage.IStorageService.StorePart"
	ARpcNameStorageIStorageServiceCompleteMultipart         = "storage.IStorageService.CompleteMultipartStore"
	ARpcNameStorageIStorageServiceAbortMultipart            = "storage.IStorageService.AbortMultiPartStore"
	ARpcNameStorageIStorageServiceGetPublicURL              = "storage.IStorageService.GetPublicURL"
	ARpcNameStorageIStorageServiceInternalStoreInit         = "storage.IStorageService._internal.StoreInit"
	ARpcNameStorageIStorageServiceInternalStoreChunk        = "storage.IStorageService._internal.StoreChunk"
	ARpcNameStorageIStorageServiceInternalStoreFinish       = "storage.IStorageService._internal.StoreFinish"
	ARpcNameStorageIStorageServiceInternalStoreAbort        = "storage.IStorageService._internal.StoreAbort"
	ARpcNameStorageIStorageServiceInternalLoadChunk         = "storage.IStorageService._internal.LoadChunk"
	defaultARpcStorageReadChunkSize                   int64 = 1 << 20 // 1 MiB
	defaultARpcStorageWriteChunkSize                  int   = 1 << 20 // 1 MiB
)

type IStorageServiceListProvidersResult struct {
	Val0 []string
}

type IStorageServiceListMetaArgs struct {
	Val0 string
	Val1 int64
	Val2 int64
}

type IStorageServiceListMetaResult struct {
	Val0 model.PaginationResult[storage.FileMeta]
	Val1 errcode.UnmarshalError
}

type IStorageServiceMetaArgs struct {
	Val0 string
}

type IStorageServiceMetaResult struct {
	Val0 storage.FileMeta
	Val1 errcode.UnmarshalError
}

type IStorageServiceDeleteArgs struct {
	Val0 string
}

type IStorageServiceDeleteResult struct {
	Val0 errcode.UnmarshalError
}

type IStorageServiceInitMultipartStoreArgs struct {
	Val0 string
	Val1 string
	Val2 storage.FileMeta
}

type IStorageServiceInitMultipartStoreResult struct {
	Val0 string
	Val1 string
	Val2 errcode.UnmarshalError
}

type IStorageServiceStorePartArgs struct {
	Val0 string
	Val1 int
	Val2 []byte
}

type IStorageServiceStorePartResult struct {
	Val0 errcode.UnmarshalError
}

type IStorageServiceCompleteMultipartStoreArgs struct {
	Val0 string
}

type IStorageServiceCompleteMultipartStoreResult struct {
	Val0 errcode.UnmarshalError
}

type IStorageServiceAbortMultiPartStoreArgs struct {
	Val0 string
}

type IStorageServiceAbortMultiPartStoreResult struct {
	Val0 errcode.UnmarshalError
}

type IStorageServiceGetPublicURLArgs struct {
	Val0 string
}

type IStorageServiceGetPublicURLResult struct {
	Val0 string
	Val1 errcode.UnmarshalError
}

type IStorageServiceInternalStoreInitArgs struct {
	Val0 string
	Val1 string
	Val2 storage.FileMeta
}

type IStorageServiceInternalStoreInitResult struct {
	Val0 string
	Val1 errcode.UnmarshalError
}

type IStorageServiceInternalStoreChunkArgs struct {
	Val0 string
	Val1 []byte
}

type IStorageServiceInternalStoreChunkResult struct {
	Val0 errcode.UnmarshalError
}

type IStorageServiceInternalStoreFinishArgs struct {
	Val0 string
}

type IStorageServiceInternalStoreFinishResult struct {
	Val0 string
	Val1 errcode.UnmarshalError
}

type IStorageServiceInternalStoreAbortArgs struct {
	Val0 string
}

type IStorageServiceInternalStoreAbortResult struct {
	Val0 errcode.UnmarshalError
}

type IStorageServiceInternalLoadChunkArgs struct {
	Val0 string
	Val1 int64
	Val2 int64
}

type IStorageServiceInternalLoadChunkResult struct {
	Val0 []byte
	Val1 errcode.UnmarshalError
}

type arpcClientIStorageService struct {
	client  sarpc.Client `aperture:""`
	timeout time.Duration
	log     logger.ILogger `aperture:""`
}

func NewARpcIStorageService(client sarpc.Client) storage.IStorageService {
	return &arpcClientIStorageService{
		client:  client,
		timeout: time.Second * 5,
	}
}

func NewARpcIStorageServiceWithTimeout(client sarpc.Client, timeout time.Duration) storage.IStorageService {
	return &arpcClientIStorageService{
		client:  client,
		timeout: timeout,
	}
}

func (r *arpcClientIStorageService) SrvImplName() scene.ImplName {
	return storage.Lens.ImplName("IStorageService", "arpc")
}

func (r *arpcClientIStorageService) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IStorageService", "arpc")
}

func (r *arpcClientIStorageService) Setup() error {
	return nil
}

func (r *arpcClientIStorageService) ListProviders() []string {
	var resp IStorageServiceListProvidersResult
	err := r.client.Call(ARpcNameStorageIStorageServiceListProviders, &struct{}{}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceListProviders, "err", err)
		return nil
	}
	return resp.Val0
}

func (r *arpcClientIStorageService) ListMeta(provider string, offset, limit int64) (model.PaginationResult[storage.FileMeta], error) {
	var resp IStorageServiceListMetaResult
	err := r.client.Call(ARpcNameStorageIStorageServiceListMeta, &IStorageServiceListMetaArgs{
		Val0: provider,
		Val1: offset,
		Val2: limit,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceListMeta, "err", err)
		return model.PaginationResult[storage.FileMeta]{}, err
	}
	return resp.Val0, resp.Val1.Error
}

func (r *arpcClientIStorageService) Meta(fileId storage.FileID) (storage.FileMeta, error) {
	var resp IStorageServiceMetaResult
	err := r.client.Call(ARpcNameStorageIStorageServiceMeta, &IStorageServiceMetaArgs{
		Val0: string(fileId),
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceMeta, "err", err)
		return storage.FileMeta{}, err
	}
	return resp.Val0, resp.Val1.Error
}

func (r *arpcClientIStorageService) Load(fileId storage.FileID, offset, length int64) (io.ReadCloser, error) {
	if length <= 0 {
		return nil, storage.ErrInvalidLength
	}
	return &remoteChunkReader{
		client:    r,
		fileId:    fileId,
		offset:    offset,
		remaining: length,
	}, nil
}

func (r *arpcClientIStorageService) LoadAll(fileId storage.FileID) (io.ReadCloser, error) {
	meta, err := r.Meta(fileId)
	if err != nil {
		return nil, err
	}
	return &remoteChunkReader{
		client:    r,
		fileId:    fileId,
		offset:    0,
		remaining: meta.ContentLength,
	}, nil
}

func (r *arpcClientIStorageService) Delete(fileId storage.FileID) error {
	var resp IStorageServiceDeleteResult
	err := r.client.Call(ARpcNameStorageIStorageServiceDelete, &IStorageServiceDeleteArgs{
		Val0: string(fileId),
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceDelete, "err", err)
		return err
	}
	return resp.Val0.Error
}

func (r *arpcClientIStorageService) Store(data io.Reader, meta storage.FileMeta) (storage.FileID, error) {
	return r.StoreAt("", "", data, meta)
}

func (r *arpcClientIStorageService) StoreAt(provider, identifier string, data io.Reader, meta storage.FileMeta) (storage.FileID, error) {
	uploadId, err := r.remoteStoreInit(provider, identifier, meta)
	if err != nil {
		return "", err
	}
	buffer := make([]byte, defaultARpcStorageWriteChunkSize)
	for {
		n, readErr := data.Read(buffer)
		if n > 0 {
			if chunkErr := r.remoteStoreChunk(uploadId, buffer[:n]); chunkErr != nil {
				_ = r.remoteStoreAbort(uploadId)
				return "", chunkErr
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			_ = r.remoteStoreAbort(uploadId)
			return "", readErr
		}
	}
	fileId, finishErr := r.remoteStoreFinish(uploadId)
	if finishErr != nil {
		_ = r.remoteStoreAbort(uploadId)
		return "", finishErr
	}
	return fileId, nil
}

func (r *arpcClientIStorageService) InitMultipartStore(provider, identifier string, meta storage.FileMeta) (storage.FileID, string, error) {
	var resp IStorageServiceInitMultipartStoreResult
	err := r.client.Call(ARpcNameStorageIStorageServiceInitMultipartStore, &IStorageServiceInitMultipartStoreArgs{
		Val0: provider,
		Val1: identifier,
		Val2: meta,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceInitMultipartStore, "err", err)
		return "", "", err
	}
	return storage.FileID(resp.Val0), resp.Val1, resp.Val2.Error
}

func (r *arpcClientIStorageService) StorePart(uploadId string, partNumber int, data io.Reader) error {
	return r.StorePartReader(uploadId, partNumber, data)
}

func (r *arpcClientIStorageService) StorePartReader(uploadId string, partNumber int, data io.Reader) error {
	raw, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	var resp IStorageServiceStorePartResult
	callErr := r.client.Call(ARpcNameStorageIStorageServiceStorePart, &IStorageServiceStorePartArgs{
		Val0: uploadId,
		Val1: partNumber,
		Val2: raw,
	}, &resp, r.timeout)
	if callErr != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceStorePart, "err", callErr)
		return callErr
	}
	return resp.Val0.Error
}

func (r *arpcClientIStorageService) CompleteMultipartStore(uploadId string) error {
	var resp IStorageServiceCompleteMultipartStoreResult
	err := r.client.Call(ARpcNameStorageIStorageServiceCompleteMultipart, &IStorageServiceCompleteMultipartStoreArgs{
		Val0: uploadId,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceCompleteMultipart, "err", err)
		return err
	}
	return resp.Val0.Error
}

func (r *arpcClientIStorageService) AbortMultiPartStore(uploadId string) error {
	var resp IStorageServiceAbortMultiPartStoreResult
	err := r.client.Call(ARpcNameStorageIStorageServiceAbortMultipart, &IStorageServiceAbortMultiPartStoreArgs{
		Val0: uploadId,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceAbortMultipart, "err", err)
		return err
	}
	return resp.Val0.Error
}

func (r *arpcClientIStorageService) GetPublicURL(fileId storage.FileID) (string, error) {
	var resp IStorageServiceGetPublicURLResult
	err := r.client.Call(ARpcNameStorageIStorageServiceGetPublicURL, &IStorageServiceGetPublicURLArgs{
		Val0: string(fileId),
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceGetPublicURL, "err", err)
		return "", err
	}
	return resp.Val0, resp.Val1.Error
}

func (r *arpcClientIStorageService) remoteStoreInit(provider, identifier string, meta storage.FileMeta) (string, error) {
	var resp IStorageServiceInternalStoreInitResult
	err := r.client.Call(ARpcNameStorageIStorageServiceInternalStoreInit, &IStorageServiceInternalStoreInitArgs{
		Val0: provider,
		Val1: identifier,
		Val2: meta,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceInternalStoreInit, "err", err)
		return "", err
	}
	return resp.Val0, resp.Val1.Error
}

func (r *arpcClientIStorageService) remoteStoreChunk(uploadId string, data []byte) error {
	var resp IStorageServiceInternalStoreChunkResult
	err := r.client.Call(ARpcNameStorageIStorageServiceInternalStoreChunk, &IStorageServiceInternalStoreChunkArgs{
		Val0: uploadId,
		Val1: data,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceInternalStoreChunk, "err", err)
		return err
	}
	return resp.Val0.Error
}

func (r *arpcClientIStorageService) remoteStoreFinish(uploadId string) (storage.FileID, error) {
	var resp IStorageServiceInternalStoreFinishResult
	err := r.client.Call(ARpcNameStorageIStorageServiceInternalStoreFinish, &IStorageServiceInternalStoreFinishArgs{
		Val0: uploadId,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceInternalStoreFinish, "err", err)
		return "", err
	}
	return storage.FileID(resp.Val0), resp.Val1.Error
}

func (r *arpcClientIStorageService) remoteStoreAbort(uploadId string) error {
	var resp IStorageServiceInternalStoreAbortResult
	err := r.client.Call(ARpcNameStorageIStorageServiceInternalStoreAbort, &IStorageServiceInternalStoreAbortArgs{
		Val0: uploadId,
	}, &resp, r.timeout)
	if err != nil {
		return err
	}
	return resp.Val0.Error
}

func (r *arpcClientIStorageService) loadChunk(fileId storage.FileID, offset, length int64) ([]byte, error) {
	var resp IStorageServiceInternalLoadChunkResult
	err := r.client.Call(ARpcNameStorageIStorageServiceInternalLoadChunk, &IStorageServiceInternalLoadChunkArgs{
		Val0: string(fileId),
		Val1: offset,
		Val2: length,
	}, &resp, r.timeout)
	if err != nil {
		r.log.ErrorW("remote call error", "method", ARpcNameStorageIStorageServiceInternalLoadChunk, "err", err)
		return nil, err
	}
	return resp.Val0, resp.Val1.Error
}

type remoteChunkReader struct {
	client    *arpcClientIStorageService
	fileId    storage.FileID
	offset    int64
	remaining int64
	pos       int64
	buffer    []byte
	bufferPos int
	closed    bool
}

func (r *remoteChunkReader) Read(p []byte) (int, error) {
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	if len(p) == 0 {
		return 0, nil
	}
	if r.remaining == 0 {
		return 0, io.EOF
	}
	if len(r.buffer)-r.bufferPos == 0 {
		chunkSize := int64(len(p))
		if chunkSize < defaultARpcStorageReadChunkSize {
			chunkSize = defaultARpcStorageReadChunkSize
		}
		if chunkSize > r.remaining {
			chunkSize = r.remaining
		}
		raw, err := r.client.loadChunk(r.fileId, r.offset+r.pos, chunkSize)
		if err != nil {
			return 0, err
		}
		if len(raw) == 0 {
			return 0, io.EOF
		}
		r.buffer = raw
		r.bufferPos = 0
	}
	n := copy(p, r.buffer[r.bufferPos:])
	r.bufferPos += n
	r.pos += int64(n)
	r.remaining -= int64(n)
	if r.remaining == 0 && n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (r *remoteChunkReader) Close() error {
	r.closed = true
	r.buffer = nil
	return nil
}

type ARpcServerIStorageService struct {
	srv storage.IStorageService `aperture:""`

	storeMu       sync.Mutex
	storeSessions map[string]*storeSession
}

type storeSession struct {
	provider   string
	identifier string
	meta       storage.FileMeta
	path       string
	file       *os.File
}

func NewARpcServerIStorageService(srv storage.IStorageService) *ARpcServerIStorageService {
	return &ARpcServerIStorageService{
		srv:           srv,
		storeSessions: make(map[string]*storeSession),
	}
}

func HandleIStorageService(srv storage.IStorageService, handler arpc.Handler) {
	server := NewARpcServerIStorageService(srv)
	handler.Handle(ARpcNameStorageIStorageServiceListProviders, server.ListProviders)
	handler.Handle(ARpcNameStorageIStorageServiceListMeta, server.ListMeta)
	handler.Handle(ARpcNameStorageIStorageServiceMeta, server.Meta)
	handler.Handle(ARpcNameStorageIStorageServiceDelete, server.Delete)
	handler.Handle(ARpcNameStorageIStorageServiceInitMultipartStore, server.InitMultipartStore)
	handler.Handle(ARpcNameStorageIStorageServiceStorePart, server.StorePart)
	handler.Handle(ARpcNameStorageIStorageServiceCompleteMultipart, server.CompleteMultipartStore)
	handler.Handle(ARpcNameStorageIStorageServiceAbortMultipart, server.AbortMultiPartStore)
	handler.Handle(ARpcNameStorageIStorageServiceGetPublicURL, server.GetPublicURL)
	handler.Handle(ARpcNameStorageIStorageServiceInternalStoreInit, server.StoreInit)
	handler.Handle(ARpcNameStorageIStorageServiceInternalStoreChunk, server.StoreChunk)
	handler.Handle(ARpcNameStorageIStorageServiceInternalStoreFinish, server.StoreFinish)
	handler.Handle(ARpcNameStorageIStorageServiceInternalStoreAbort, server.StoreAbort)
	handler.Handle(ARpcNameStorageIStorageServiceInternalLoadChunk, server.LoadChunk)
}

func (r *ARpcServerIStorageService) ListProviders(c *arpc.Context) {
	resp := IStorageServiceListProvidersResult{Val0: r.srv.ListProviders()}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) ListMeta(c *arpc.Context) {
	var req IStorageServiceListMetaArgs
	resp := IStorageServiceListMetaResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	a0, a1 := r.srv.ListMeta(req.Val0, req.Val1, req.Val2)
	resp.Val0 = a0
	resp.Val1 = errcode.UnmarshalError{Error: a1}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) Meta(c *arpc.Context) {
	var req IStorageServiceMetaArgs
	resp := IStorageServiceMetaResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	fileId, ok := storage.ParseFileID(req.Val0)
	if !ok {
		resp.Val1 = errcode.UnmarshalError{Error: storage.ErrInvalidFileID}
		_ = c.Write(&resp)
		return
	}
	a0, a1 := r.srv.Meta(fileId)
	resp.Val0 = a0
	resp.Val1 = errcode.UnmarshalError{Error: a1}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) Delete(c *arpc.Context) {
	var req IStorageServiceDeleteArgs
	resp := IStorageServiceDeleteResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	fileId, ok := storage.ParseFileID(req.Val0)
	if !ok {
		resp.Val0 = errcode.UnmarshalError{Error: storage.ErrInvalidFileID}
		_ = c.Write(&resp)
		return
	}
	resp.Val0 = errcode.UnmarshalError{Error: r.srv.Delete(fileId)}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) InitMultipartStore(c *arpc.Context) {
	var req IStorageServiceInitMultipartStoreArgs
	resp := IStorageServiceInitMultipartStoreResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val2 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	a0, a1, a2 := r.srv.InitMultipartStore(req.Val0, req.Val1, req.Val2)
	resp.Val0 = string(a0)
	resp.Val1 = a1
	resp.Val2 = errcode.UnmarshalError{Error: a2}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) StorePart(c *arpc.Context) {
	var req IStorageServiceStorePartArgs
	resp := IStorageServiceStorePartResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	resp.Val0 = errcode.UnmarshalError{Error: r.srv.StorePartReader(req.Val0, req.Val1, bytes.NewReader(req.Val2))}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) CompleteMultipartStore(c *arpc.Context) {
	var req IStorageServiceCompleteMultipartStoreArgs
	resp := IStorageServiceCompleteMultipartStoreResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	resp.Val0 = errcode.UnmarshalError{Error: r.srv.CompleteMultipartStore(req.Val0)}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) AbortMultiPartStore(c *arpc.Context) {
	var req IStorageServiceAbortMultiPartStoreArgs
	resp := IStorageServiceAbortMultiPartStoreResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	resp.Val0 = errcode.UnmarshalError{Error: r.srv.AbortMultiPartStore(req.Val0)}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) GetPublicURL(c *arpc.Context) {
	var req IStorageServiceGetPublicURLArgs
	resp := IStorageServiceGetPublicURLResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	fileId, ok := storage.ParseFileID(req.Val0)
	if !ok {
		resp.Val1 = errcode.UnmarshalError{Error: storage.ErrInvalidFileID}
		_ = c.Write(&resp)
		return
	}
	a0, a1 := r.srv.GetPublicURL(fileId)
	resp.Val0 = a0
	resp.Val1 = errcode.UnmarshalError{Error: a1}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) StoreInit(c *arpc.Context) {
	var req IStorageServiceInternalStoreInitArgs
	resp := IStorageServiceInternalStoreInitResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	file, err := os.CreateTemp("", "scene-arpc-storage-*")
	if err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	uploadId := uuid.NewString()
	r.storeMu.Lock()
	r.storeSessions[uploadId] = &storeSession{
		provider:   req.Val0,
		identifier: req.Val1,
		meta:       req.Val2,
		path:       file.Name(),
		file:       file,
	}
	r.storeMu.Unlock()
	resp.Val0 = uploadId
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) StoreChunk(c *arpc.Context) {
	var req IStorageServiceInternalStoreChunkArgs
	resp := IStorageServiceInternalStoreChunkResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	r.storeMu.Lock()
	sess, ok := r.storeSessions[req.Val0]
	r.storeMu.Unlock()
	if !ok {
		resp.Val0 = errcode.UnmarshalError{Error: storage.ErrUploadSessionNotFound}
		_ = c.Write(&resp)
		return
	}
	if _, err := sess.file.Write(req.Val1); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) StoreFinish(c *arpc.Context) {
	var req IStorageServiceInternalStoreFinishArgs
	resp := IStorageServiceInternalStoreFinishResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	r.storeMu.Lock()
	sess, ok := r.storeSessions[req.Val0]
	if ok {
		delete(r.storeSessions, req.Val0)
	}
	r.storeMu.Unlock()
	if !ok {
		resp.Val1 = errcode.UnmarshalError{Error: storage.ErrUploadSessionNotFound}
		_ = c.Write(&resp)
		return
	}
	defer func() {
		_ = sess.file.Close()
		_ = os.Remove(sess.path)
	}()
	if _, err := sess.file.Seek(0, io.SeekStart); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	fileId, err := r.srv.StoreAt(sess.provider, sess.identifier, sess.file, sess.meta)
	resp.Val0 = string(fileId)
	resp.Val1 = errcode.UnmarshalError{Error: err}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) StoreAbort(c *arpc.Context) {
	var req IStorageServiceInternalStoreAbortArgs
	resp := IStorageServiceInternalStoreAbortResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val0 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	r.storeMu.Lock()
	sess, ok := r.storeSessions[req.Val0]
	if ok {
		delete(r.storeSessions, req.Val0)
	}
	r.storeMu.Unlock()
	if ok {
		_ = sess.file.Close()
		_ = os.Remove(sess.path)
	}
	_ = c.Write(&resp)
}

func (r *ARpcServerIStorageService) LoadChunk(c *arpc.Context) {
	var req IStorageServiceInternalLoadChunkArgs
	resp := IStorageServiceInternalLoadChunkResult{}
	if err := c.Bind(&req); err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	fileId, ok := storage.ParseFileID(req.Val0)
	if !ok {
		resp.Val1 = errcode.UnmarshalError{Error: storage.ErrInvalidFileID}
		_ = c.Write(&resp)
		return
	}
	reader, err := r.srv.Load(fileId, req.Val1, req.Val2)
	if err != nil {
		resp.Val1 = errcode.UnmarshalError{Error: err}
		_ = c.Write(&resp)
		return
	}
	data, readErr := io.ReadAll(reader)
	if closeErr := reader.Close(); closeErr != nil && readErr == nil {
		readErr = closeErr
	}
	resp.Val0 = data
	resp.Val1 = errcode.UnmarshalError{Error: readErr}
	_ = c.Write(&resp)
}

type ARpcAppIStorageService struct {
	srv storage.IStorageService `aperture:""`
}

func (r *ARpcAppIStorageService) Name() scene.ImplName {
	return storage.Lens.ImplNameNoVer("ARpcApplication")
}

func (r *ARpcAppIStorageService) RegisterService(handler arpc.Handler) error {
	HandleIStorageService(r.srv, handler)
	return nil
}

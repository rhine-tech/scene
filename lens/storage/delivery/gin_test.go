package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"github.com/stretchr/testify/require"
)

type urlTestStorageService struct {
	storage.IStorageService
	directURL  string
	directErr  error
	directKeys []storage.StorageKey
	directCtxs []context.Context
}

type dataTestLoadCall struct {
	offset int64
	length int64
}

type dataTestStorageService struct {
	storage.IStorageService
	data      []byte
	loadCalls []dataTestLoadCall
	metaCalls int
	reads     int
	closes    int
}

type dataTestReadCloser struct {
	service *dataTestStorageService
	reader  io.Reader
	closed  bool
}

func (r *dataTestReadCloser) Read(p []byte) (int, error) {
	r.service.reads++
	return r.reader.Read(p)
}

func (r *dataTestReadCloser) Close() error {
	if !r.closed {
		r.closed = true
		r.service.closes++
	}
	return nil
}

func (s *dataTestStorageService) Meta(_ context.Context, storageKey storage.StorageKey) (storage.FileMeta, error) {
	s.metaCalls++
	return storage.FileMeta{
		StorageKey:       storageKey,
		OriginalFilename: "download",
		ContentLength:    int64(len(s.data)),
	}, nil
}

func (s *dataTestStorageService) Load(_ context.Context, _ storage.StorageKey, offset, length int64) (io.ReadCloser, error) {
	s.loadCalls = append(s.loadCalls, dataTestLoadCall{offset: offset, length: length})
	if offset < 0 || length <= 0 || offset > int64(len(s.data)) {
		return nil, storage.ErrInvalidOffset
	}
	end := offset + length
	if end > int64(len(s.data)) {
		end = int64(len(s.data))
	}
	return &dataTestReadCloser{service: s, reader: bytes.NewReader(s.data[offset:end])}, nil
}

func (s *urlTestStorageService) GetDirectURL(ctx context.Context, storageKey storage.StorageKey) (string, error) {
	s.directCtxs = append(s.directCtxs, ctx)
	s.directKeys = append(s.directKeys, storageKey)
	return s.directURL, s.directErr
}

func TestGetURLDefaultsToProxyUsingMountedRoutePrefix(t *testing.T) {
	srv := &urlTestStorageService{}
	router := newURLTestRouter(srv)

	resp := performURLRequest(t, router, "/api/storage/url/local.default/dir/file.txt")

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "/api/storage/data/local.default/dir/file.txt", responseData(t, resp))
	require.Empty(t, srv.directKeys)
}

func TestGetURLProxyModeDoesNotCallService(t *testing.T) {
	srv := &urlTestStorageService{directErr: errcode.InternalError}
	router := newURLTestRouter(srv)

	resp := performURLRequest(t, router, "/api/storage/url/local.default/file.txt?mode=proxy")

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "/api/storage/data/local.default/file.txt", responseData(t, resp))
	require.Empty(t, srv.directKeys)
}

func TestGetURLProxyModeEscapesIdentifier(t *testing.T) {
	srv := &urlTestStorageService{}
	router := newURLTestRouter(srv)

	resp := performURLRequest(t, router, "/api/storage/url/local.default/a%20file%3F.txt?mode=proxy")

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "/api/storage/data/local.default/a%20file%3F.txt", responseData(t, resp))
}

func TestGetURLDirectModeCallsService(t *testing.T) {
	srv := &urlTestStorageService{directURL: "https://storage.example/file.txt?signature=test"}
	router := newURLTestRouter(srv)

	resp := performURLRequest(t, router, "/api/storage/url/s3.default/dir/file.txt?mode=direct")

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, srv.directURL, responseData(t, resp))
	require.Equal(t, []storage.StorageKey{"s3.default://dir/file.txt"}, srv.directKeys)
}

func TestGetURLDirectModePassesRequestContext(t *testing.T) {
	type contextKey struct{}
	srv := &urlTestStorageService{directURL: "https://storage.example/file.txt"}
	router := newURLTestRouter(srv)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/storage/url/s3.default/file.txt?mode=direct", nil)
	request = request.WithContext(context.WithValue(request.Context(), contextKey{}, "request-value"))

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, srv.directCtxs, 1)
	require.Equal(t, "request-value", srv.directCtxs[0].Value(contextKey{}))
}

func TestGetURLRejectsUnknownMode(t *testing.T) {
	srv := &urlTestStorageService{}
	router := newURLTestRouter(srv)

	resp := performURLRequest(t, router, "/api/storage/url/local.default/file.txt?mode=unknown")

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, errcode.ParameterError.Code, responseCode(t, resp))
	require.Empty(t, srv.directKeys)
}

func TestGetDataSequentialHTTPReadsUseOneProviderReader(t *testing.T) {
	data := bytes.Repeat([]byte("stream-content-"), 160000)
	srv := &dataTestStorageService{data: data}
	router := newDataTestRouter(srv)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/storage/data/local.default/big.bin", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/octet-stream", recorder.Header().Get("Content-Type"))
	require.Equal(t, strconv.Itoa(len(data)), recorder.Header().Get("Content-Length"))
	require.Equal(t, data, recorder.Body.Bytes())
	require.Greater(t, srv.reads, 1)
	require.Equal(t, []dataTestLoadCall{{offset: 0, length: int64(len(data))}}, srv.loadCalls)
	require.Equal(t, 1, srv.closes)
}

func TestGetDataRequiresDownloadPermission(t *testing.T) {
	srv := &dataTestStorageService{data: []byte("private")}
	router := newProtectedDataTestRouter(srv)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/storage/data/local.default/private.bin", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Equal(t, permission.ErrPermissionDenied.Code, responseCode(t, recorder))
	require.Contains(t, responseMessage(t, recorder), storage.PermFileDownload.String())
	require.Zero(t, srv.metaCalls)
	require.Empty(t, srv.loadCalls)
}

func TestGetDataSingleRangeUsesOneProviderReader(t *testing.T) {
	data := bytes.Repeat([]byte("0123456789"), 200000)
	srv := &dataTestStorageService{data: data}
	router := newDataTestRouter(srv)
	const start = int64(1_048_570)
	const end = int64(1_081_337)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/storage/data/local.default/big.bin", nil)
	request.Header.Set("Range", "bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10))

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusPartialContent, recorder.Code)
	require.Equal(t, data[start:end+1], recorder.Body.Bytes())
	require.Equal(t, []dataTestLoadCall{{offset: start, length: int64(len(data)) - start}}, srv.loadCalls)
	require.Equal(t, 1, srv.closes)
}

func TestGetDataHeadDoesNotOpenProviderReader(t *testing.T) {
	data := bytes.Repeat([]byte("head"), 1024)
	srv := &dataTestStorageService{data: data}
	router := newDataTestRouter(srv)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodHead, "/api/storage/data/local.default/file.bin", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, strconv.Itoa(len(data)), recorder.Header().Get("Content-Length"))
	require.Empty(t, recorder.Body.Bytes())
	require.Empty(t, srv.loadCalls)
	require.Zero(t, srv.closes)
}

func TestGetDataMultipleRangesOpenOncePerRange(t *testing.T) {
	data := bytes.Repeat([]byte("0123456789"), 100)
	srv := &dataTestStorageService{data: data}
	router := newDataTestRouter(srv)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/storage/data/local.default/file.bin", nil)
	request.Header.Set("Range", "bytes=0-2,10-12")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusPartialContent, recorder.Code)
	require.Equal(t, []dataTestLoadCall{
		{offset: 0, length: int64(len(data))},
		{offset: 10, length: int64(len(data)) - 10},
	}, srv.loadCalls)
	require.Equal(t, 2, srv.closes)
}

func newURLTestRouter(srv storage.IStorageService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app := &appContext{srv: srv}
	router.GET("/api/storage"+urlRoutePath, sgin.Handle(app, new(getURLRequest)))
	return router
}

func newDataTestRouter(srv storage.IStorageService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app := &appContext{srv: srv}
	router.GET("/api/storage"+dataRoutePath, sgin.Handle(app, new(getDataRequest)))
	router.HEAD("/api/storage"+dataRoutePath, sgin.Handle(app, new(getDataRequest)))
	return router
}

func newProtectedDataTestRouter(srv storage.IStorageService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app := &appContext{srv: srv}
	appRouter := sgin.NewAppRouter(app, router.Group("/api/storage"), nil)
	appRouter.HandleAction(new(getDataRequest))
	return router
}

func performURLRequest(t *testing.T, router http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	router.ServeHTTP(recorder, request)
	return recorder
}

func responseData(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var body model.AppResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	data, ok := body.Data.(string)
	require.True(t, ok)
	return data
}

func responseCode(t *testing.T, response *httptest.ResponseRecorder) int {
	t.Helper()
	var body model.AppResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	return body.Code
}

func responseMessage(t *testing.T, response *httptest.ResponseRecorder) string {
	t.Helper()
	var body model.AppResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	return body.Msg
}

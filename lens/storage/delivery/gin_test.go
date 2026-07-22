package delivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/errcode"
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
}

func (s *urlTestStorageService) GetDirectURL(storageKey storage.StorageKey) (string, error) {
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

func TestGetURLRejectsUnknownMode(t *testing.T) {
	srv := &urlTestStorageService{}
	router := newURLTestRouter(srv)

	resp := performURLRequest(t, router, "/api/storage/url/local.default/file.txt?mode=unknown")

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Equal(t, errcode.ParameterError.Code, responseCode(t, resp))
	require.Empty(t, srv.directKeys)
}

func newURLTestRouter(srv storage.IStorageService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app := &appContext{srv: srv}
	router.GET("/api/storage"+urlRoutePath, sgin.Handle(app, new(getURLRequest)))
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

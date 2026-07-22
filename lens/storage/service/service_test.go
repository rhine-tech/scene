package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	storageapi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	"github.com/stretchr/testify/require"
)

type testLogger struct{}

func (testLogger) Debug(args ...interface{})                           {}
func (testLogger) Debugf(format string, args ...interface{})           {}
func (testLogger) DebugW(message string, keysAndValues ...interface{}) {}
func (testLogger) Info(args ...interface{})                            {}
func (testLogger) Infof(format string, args ...interface{})            {}
func (testLogger) InfoW(message string, keysAndValues ...interface{})  {}
func (testLogger) Warn(args ...interface{})                            {}
func (testLogger) Warnf(format string, args ...interface{})            {}
func (testLogger) WarnW(message string, keysAndValues ...interface{})  {}
func (testLogger) Error(args ...interface{})                           {}
func (testLogger) Errorf(format string, args ...interface{})           {}
func (testLogger) ErrorW(message string, keysAndValues ...interface{}) {}
func (l testLogger) WithPrefix(prefix string) logger.ILogger           { return l }
func (testLogger) SetLogLevel(level logger.LogLevel)                   {}
func (l testLogger) WithOptions(opts ...logger.Option) logger.ILogger  { return l }

type testMetaRepo struct {
	storeErr error
	loadErr  error
	listErr  error
}

func (r *testMetaRepo) ImplName() scene.ImplName {
	return storageapi.Lens.ImplName("IFileMetaRepository", "test")
}

func (r *testMetaRepo) Store(context.Context, storageapi.FileMeta) error {
	return r.storeErr
}

func (r *testMetaRepo) Load(context.Context, storageapi.StorageKey) (storageapi.FileMeta, error) {
	return storageapi.FileMeta{}, r.loadErr
}

func (r *testMetaRepo) Delete(context.Context, storageapi.StorageKey) error {
	return nil
}

func (r *testMetaRepo) List(
	context.Context,
	string,
	int64,
	int64,
) (model.PaginationResult[storageapi.FileMeta], error) {
	return model.PaginationResult[storageapi.FileMeta]{}, r.listErr
}

type testSessionTracker struct {
	saveErr error
}

func (t *testSessionTracker) ImplName() scene.ImplName {
	return storageapi.Lens.ImplName("IUploadSessionTracker", "test")
}

func (t *testSessionTracker) Save(uploadId string, session storageapi.UploadSession) error {
	return t.saveErr
}

func (t *testSessionTracker) Get(uploadId string) (storageapi.UploadSession, error) {
	return storageapi.UploadSession{}, storageapi.ErrUploadSessionNotFound
}

func (t *testSessionTracker) Delete(uploadId string) error {
	return nil
}

type testProvider struct {
	name           string
	storeErr       error
	initErr        error
	directURLErr   error
	abortCalled    bool
	initializedKey storageapi.StorageKey
}

func (p *testProvider) ImplName() scene.ImplName {
	return storageapi.Lens.ImplName("IStorageProvider", "test")
}

func (p *testProvider) ProviderName() string {
	if p.name != "" {
		return p.name
	}
	return "local.test"
}

func (p *testProvider) HealthCheck() error { return nil }

func (p *testProvider) Meta(storageKey storageapi.StorageKey) (storageapi.FileMeta, error) {
	return storageapi.FileMeta{}, storageapi.ErrFileNotFound
}

func (p *testProvider) Load(storageKey storageapi.StorageKey, offset, length int64) (io.ReadCloser, error) {
	return nil, storageapi.ErrFileNotFound
}

func (p *testProvider) LoadAll(storageKey storageapi.StorageKey) (io.ReadCloser, error) {
	return nil, storageapi.ErrFileNotFound
}

func (p *testProvider) Store(storageKey storageapi.StorageKey, data io.Reader) error {
	_, _ = io.Copy(io.Discard, data)
	return p.storeErr
}

func (p *testProvider) Delete(storageKey storageapi.StorageKey) error {
	return nil
}

func (p *testProvider) InitMultipartStore(storageKey storageapi.StorageKey) (string, error) {
	p.initializedKey = storageKey
	return "upload-test", p.initErr
}

func (p *testProvider) StoreMultipart(uploadId string, partNumber int, data io.Reader) error {
	return nil
}

func (p *testProvider) CompleteMultipart(uploadId string) error {
	return nil
}

func (p *testProvider) AbortMultipart(uploadId string) error {
	p.abortCalled = true
	return nil
}

func (p *testProvider) GetDirectURL(storageKey storageapi.StorageKey) (string, error) {
	return "", p.directURLErr
}

func newTestService(repo *testMetaRepo, tracker *testSessionTracker, provider *testProvider) *StorageService {
	srv := NewStorageService(repo, tracker, "local.test", provider)
	srv.log = testLogger{}
	return srv
}

func TestListProvidersSortedAndReturnsCopy(t *testing.T) {
	srv := NewStorageService(
		&testMetaRepo{},
		&testSessionTracker{},
		"local.default",
		&testProvider{name: "s3.archive"},
		&testProvider{name: "local.default"},
		&testProvider{name: "s3.default"},
	)
	want := []string{"local.default", "s3.archive", "s3.default"}

	got := srv.ListProviders()
	require.Equal(t, want, got)

	got[0] = "modified"
	require.Equal(t, want, srv.ListProviders())
}

func TestListProvidersConcurrentFirstRead(t *testing.T) {
	srv := NewStorageService(
		&testMetaRepo{},
		&testSessionTracker{},
		"local.default",
		&testProvider{name: "s3.archive"},
		&testProvider{name: "local.default"},
		&testProvider{name: "s3.default"},
	)
	want := []string{"local.default", "s3.archive", "s3.default"}

	for range 16 {
		t.Run("reader", func(t *testing.T) {
			t.Parallel()
			require.Equal(t, want, srv.ListProviders())
		})
	}
}

func TestStoreAtWrapsMetaRepositoryError(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{storeErr: errors.New("sql: hidden details")},
		&testSessionTracker{},
		&testProvider{},
	)

	_, err := srv.StoreAt("local.test", "covers/work.jpg", bytes.NewReader([]byte("data")), storageapi.FileMeta{})
	require.ErrorIs(t, err, storageapi.ErrFailToStore)
	require.NotContains(t, err.Error(), "sql:")
}

func TestStoreAtPreservesStorageKeyExists(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{},
		&testSessionTracker{},
		&testProvider{storeErr: storageapi.ErrStorageKeyExists},
	)

	_, err := srv.StoreAt("local.test", "covers/work.jpg", bytes.NewReader([]byte("data")), storageapi.FileMeta{})
	require.ErrorIs(t, err, storageapi.ErrStorageKeyExists)
}

func TestInitMultipartStoreWrapsMetaLoadError(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{loadErr: errors.New("sql: hidden details")},
		&testSessionTracker{},
		&testProvider{},
	)

	_, _, err := srv.InitMultipartStore("local.test", "covers/work.jpg", storageapi.FileMeta{})
	require.ErrorIs(t, err, storageapi.ErrInitPartUploadFailed)
	require.NotContains(t, err.Error(), "sql:")
}

func TestInitMultipartStoreWrapsSessionSaveError(t *testing.T) {
	provider := &testProvider{}
	srv := newTestService(
		&testMetaRepo{loadErr: storageapi.ErrMetaNotFound},
		&testSessionTracker{saveErr: errors.New("redis: hidden details")},
		provider,
	)

	_, _, err := srv.InitMultipartStore("local.test", "covers/work.jpg", storageapi.FileMeta{})
	require.ErrorIs(t, err, storageapi.ErrInitPartUploadFailed)
	require.NotContains(t, err.Error(), "redis:")
	require.True(t, provider.abortCalled)
}

func TestGetDirectURLWrapsProviderError(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{},
		&testSessionTracker{},
		&testProvider{directURLErr: errors.New("s3: hidden details")},
	)

	_, err := srv.GetDirectURL("local.test://covers/work.jpg")
	require.ErrorIs(t, err, storageapi.ErrGetDirectURLFailed)
	require.NotContains(t, err.Error(), "s3:")
}

func TestGetDirectURLPreservesUnsupportedError(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{},
		&testSessionTracker{},
		&testProvider{directURLErr: storageapi.ErrDirectURLUnsupported},
	)

	_, err := srv.GetDirectURL("local.test://covers/work.jpg")
	require.ErrorIs(t, err, storageapi.ErrDirectURLUnsupported)
}

func TestServiceRejectsInvalidStorageKey(t *testing.T) {
	srv := newTestService(&testMetaRepo{}, &testSessionTracker{}, &testProvider{})

	_, err := srv.GetDirectURL("local.test:///../secret")
	require.ErrorIs(t, err, storageapi.ErrInvalidStorageKey)
}

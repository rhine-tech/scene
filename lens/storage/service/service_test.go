package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
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
	storeErr  error
	loadErr   error
	listErr   error
	deleteErr error
	storeFn   func(context.Context, storageapi.FileMeta) error
	loadFn    func(context.Context, storageapi.StorageKey) (storageapi.FileMeta, error)
}

func (r *testMetaRepo) ImplName() scene.ImplName {
	return storageapi.Lens.ImplName("IFileMetaRepository", "test")
}

func (r *testMetaRepo) Store(ctx context.Context, meta storageapi.FileMeta) error {
	if r.storeFn != nil {
		return r.storeFn(ctx, meta)
	}
	return r.storeErr
}

func (r *testMetaRepo) Load(ctx context.Context, key storageapi.StorageKey) (storageapi.FileMeta, error) {
	if r.loadFn != nil {
		return r.loadFn(ctx, key)
	}
	return storageapi.FileMeta{}, r.loadErr
}

func (r *testMetaRepo) Delete(context.Context, storageapi.StorageKey) error {
	return r.deleteErr
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
	saveErr   error
	deleteErr error
	saveFn    func(context.Context, string, storageapi.UploadSession) error
	getFn     func(context.Context, string) (storageapi.UploadSession, error)
	deleteFn  func(context.Context, string) error
}

func (t *testSessionTracker) ImplName() scene.ImplName {
	return storageapi.Lens.ImplName("IUploadSessionTracker", "test")
}

func (t *testSessionTracker) Save(ctx context.Context, uploadID string, session storageapi.UploadSession) error {
	if t.saveFn != nil {
		return t.saveFn(ctx, uploadID, session)
	}
	return t.saveErr
}

func (t *testSessionTracker) Get(ctx context.Context, uploadID string) (storageapi.UploadSession, error) {
	if t.getFn != nil {
		return t.getFn(ctx, uploadID)
	}
	return storageapi.UploadSession{}, storageapi.ErrUploadSessionNotFound
}

func (t *testSessionTracker) Delete(ctx context.Context, uploadID string) error {
	if t.deleteFn != nil {
		return t.deleteFn(ctx, uploadID)
	}
	return t.deleteErr
}

type testProvider struct {
	name           string
	storeErr       error
	initErr        error
	directURLErr   error
	abortCalled    bool
	metaCalled     bool
	initializedKey storageapi.StorageKey
	storeFn        func(context.Context, storageapi.StorageKey, io.Reader) error
	abortFn        func(context.Context, string) error
	completeFn     func(context.Context, string) error
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

func (p *testProvider) HealthCheck(context.Context) error { return nil }

func (p *testProvider) Meta(context.Context, storageapi.StorageKey) (storageapi.FileMeta, error) {
	p.metaCalled = true
	return storageapi.FileMeta{}, storageapi.ErrFileNotFound
}

func (p *testProvider) Load(context.Context, storageapi.StorageKey, int64, int64) (io.ReadCloser, error) {
	return nil, storageapi.ErrFileNotFound
}

func (p *testProvider) LoadAll(context.Context, storageapi.StorageKey) (io.ReadCloser, error) {
	return nil, storageapi.ErrFileNotFound
}

func (p *testProvider) Store(ctx context.Context, key storageapi.StorageKey, data io.Reader) error {
	if p.storeFn != nil {
		return p.storeFn(ctx, key, data)
	}
	_, _ = io.Copy(io.Discard, data)
	return p.storeErr
}

func (p *testProvider) Delete(context.Context, storageapi.StorageKey) error {
	return nil
}

func (p *testProvider) InitMultipartStore(_ context.Context, storageKey storageapi.StorageKey) (string, error) {
	p.initializedKey = storageKey
	return "upload-test", p.initErr
}

func (p *testProvider) StoreMultipart(context.Context, string, int, io.Reader) error {
	return nil
}

func (p *testProvider) CompleteMultipart(ctx context.Context, uploadID string) error {
	if p.completeFn != nil {
		return p.completeFn(ctx, uploadID)
	}
	return nil
}

func (p *testProvider) AbortMultipart(ctx context.Context, uploadID string) error {
	p.abortCalled = true
	if p.abortFn != nil {
		return p.abortFn(ctx, uploadID)
	}
	return nil
}

func (p *testProvider) GetDirectURL(context.Context, storageapi.StorageKey) (string, error) {
	return "", p.directURLErr
}

func newTestService(repo *testMetaRepo, tracker *testSessionTracker, provider *testProvider) *StorageService {
	srv := NewStorageService(repo, tracker, "local.test", provider)
	srv.log = testLogger{}
	return srv
}

func requireServiceErrcode(t *testing.T, err error, want *errcode.Error) {
	t.Helper()
	var actual *errcode.Error
	require.ErrorAs(t, err, &actual)
	require.ErrorIs(t, err, want)
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

	_, err := srv.StoreAt(context.Background(), "local.test", "covers/work.jpg", bytes.NewReader([]byte("data")), storageapi.FileMeta{})
	require.ErrorIs(t, err, storageapi.ErrFailToStore)
	require.NotContains(t, err.Error(), "sql:")
}

func TestStoreAtPreservesStorageKeyExists(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{},
		&testSessionTracker{},
		&testProvider{storeErr: storageapi.ErrStorageKeyExists},
	)

	_, err := srv.StoreAt(context.Background(), "local.test", "covers/work.jpg", bytes.NewReader([]byte("data")), storageapi.FileMeta{})
	require.ErrorIs(t, err, storageapi.ErrStorageKeyExists)
}

func TestInitMultipartStoreWrapsMetaLoadError(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{loadErr: errors.New("sql: hidden details")},
		&testSessionTracker{},
		&testProvider{},
	)

	_, _, err := srv.InitMultipartStore(context.Background(), "local.test", "covers/work.jpg", storageapi.FileMeta{})
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

	_, _, err := srv.InitMultipartStore(context.Background(), "local.test", "covers/work.jpg", storageapi.FileMeta{})
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

	_, err := srv.GetDirectURL(context.Background(), "local.test://covers/work.jpg")
	require.ErrorIs(t, err, storageapi.ErrGetDirectURLFailed)
	require.NotContains(t, err.Error(), "s3:")
}

func TestGetDirectURLPreservesUnsupportedError(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{},
		&testSessionTracker{},
		&testProvider{directURLErr: storageapi.ErrDirectURLUnsupported},
	)

	_, err := srv.GetDirectURL(context.Background(), "local.test://covers/work.jpg")
	require.ErrorIs(t, err, storageapi.ErrDirectURLUnsupported)
}

func TestServiceRejectsInvalidStorageKey(t *testing.T) {
	srv := newTestService(&testMetaRepo{}, &testSessionTracker{}, &testProvider{})

	_, err := srv.GetDirectURL(context.Background(), "local.test:///../secret")
	require.ErrorIs(t, err, storageapi.ErrInvalidStorageKey)
}

func TestStoreAtMapsContextErrorsToErrcode(t *testing.T) {
	tests := []struct {
		name       string
		newContext func() (context.Context, context.CancelFunc)
		contextErr error
	}{
		{
			name: "canceled",
			newContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			contextErr: context.Canceled,
		},
		{
			name: "deadline exceeded",
			newContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 0)
			},
			contextErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metaStored := false
			provider := &testProvider{
				storeFn: func(ctx context.Context, _ storageapi.StorageKey, _ io.Reader) error {
					return ctx.Err()
				},
			}
			srv := newTestService(
				&testMetaRepo{storeFn: func(context.Context, storageapi.FileMeta) error {
					metaStored = true
					return nil
				}},
				&testSessionTracker{},
				provider,
			)
			ctx, cancel := tt.newContext()
			defer cancel()

			_, err := srv.StoreAt(ctx, "local.test", "cancelled.bin", bytes.NewReader([]byte("data")), storageapi.FileMeta{})

			requireServiceErrcode(t, err, storageapi.ErrFailToStore)
			require.NotErrorIs(t, err, tt.contextErr)
			require.False(t, metaStored)
		})
	}
}

func TestStoreAtDoesNotMaskProviderErrorWithLaterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	provider := &testProvider{storeFn: func(context.Context, storageapi.StorageKey, io.Reader) error {
		cancel()
		return storageapi.ErrStorageKeyExists
	}}
	srv := newTestService(&testMetaRepo{}, &testSessionTracker{}, provider)

	_, err := srv.StoreAt(ctx, "local.test", "existing.bin", bytes.NewReader([]byte("data")), storageapi.FileMeta{})

	require.ErrorIs(t, err, storageapi.ErrStorageKeyExists)
	require.NotErrorIs(t, err, context.Canceled)
}

func TestStoreAtFinalizesMetadataAfterRemoteCommit(t *testing.T) {
	type contextKey struct{}
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request-value"))
	provider := &testProvider{
		storeFn: func(ctx context.Context, _ storageapi.StorageKey, data io.Reader) error {
			require.Equal(t, "request-value", ctx.Value(contextKey{}))
			_, err := io.Copy(io.Discard, data)
			cancel()
			return err
		},
	}
	metaStored := false
	srv := newTestService(
		&testMetaRepo{storeFn: func(ctx context.Context, _ storageapi.FileMeta) error {
			require.NoError(t, ctx.Err())
			require.Equal(t, "request-value", ctx.Value(contextKey{}))
			metaStored = true
			return nil
		}},
		&testSessionTracker{},
		provider,
	)

	_, err := srv.StoreAt(requestCtx, "local.test", "committed.bin", bytes.NewReader([]byte("data")), storageapi.FileMeta{})

	require.NoError(t, err)
	require.True(t, metaStored)
}

func TestMetaDoesNotFallbackToProviderAfterCancellation(t *testing.T) {
	provider := &testProvider{}
	srv := newTestService(
		&testMetaRepo{loadFn: func(ctx context.Context, _ storageapi.StorageKey) (storageapi.FileMeta, error) {
			return storageapi.FileMeta{}, ctx.Err()
		}},
		&testSessionTracker{},
		provider,
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := srv.Meta(ctx, "local.test://cancelled.bin")

	requireServiceErrcode(t, err, storageapi.ErrLoadingMeta)
	require.NotErrorIs(t, err, context.Canceled)
	require.False(t, provider.metaCalled)
}

func TestInitMultipartStoreRollsBackWithDetachedContext(t *testing.T) {
	type contextKey struct{}
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request-value"))
	rollbackCtxActive := false
	provider := &testProvider{abortFn: func(ctx context.Context, _ string) error {
		require.NoError(t, ctx.Err())
		require.Equal(t, "request-value", ctx.Value(contextKey{}))
		rollbackCtxActive = true
		return nil
	}}
	trackerDeleted := false
	tracker := &testSessionTracker{
		saveFn: func(context.Context, string, storageapi.UploadSession) error {
			cancel()
			return context.Canceled
		},
		deleteFn: func(ctx context.Context, _ string) error {
			require.NoError(t, ctx.Err())
			require.Equal(t, "request-value", ctx.Value(contextKey{}))
			trackerDeleted = true
			return nil
		},
	}
	srv := newTestService(
		&testMetaRepo{loadErr: storageapi.ErrMetaNotFound},
		tracker,
		provider,
	)

	_, _, err := srv.InitMultipartStore(requestCtx, "local.test", "cancelled.bin", storageapi.FileMeta{})

	requireServiceErrcode(t, err, storageapi.ErrInitPartUploadFailed)
	require.NotErrorIs(t, err, context.Canceled)
	require.True(t, provider.abortCalled)
	require.True(t, rollbackCtxActive)
	require.True(t, trackerDeleted)
}

func TestMultipartSessionLookupContextErrorMapsToOperationErrcode(t *testing.T) {
	tracker := &testSessionTracker{
		getFn: func(ctx context.Context, _ string) (storageapi.UploadSession, error) {
			return storageapi.UploadSession{}, ctx.Err()
		},
	}
	srv := newTestService(&testMetaRepo{}, tracker, &testProvider{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		call func() error
		want *errcode.Error
	}{
		{
			name: "store",
			call: func() error {
				return srv.StoreMultipart(ctx, "upload", 1, bytes.NewReader(nil))
			},
			want: storageapi.ErrStorePartFailed,
		},
		{
			name: "complete",
			call: func() error {
				_, err := srv.CompleteMultipart(ctx, "upload")
				return err
			},
			want: storageapi.ErrStorePartFailed,
		},
		{
			name: "abort",
			call: func() error {
				return srv.AbortMultipart(ctx, "upload")
			},
			want: storageapi.ErrFailToAbortMultipartStore,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			requireServiceErrcode(t, err, tt.want)
			require.NotErrorIs(t, err, context.Canceled)
		})
	}
}

func TestDeleteReturnsErrorWhenMetadataFinalizationFails(t *testing.T) {
	srv := newTestService(
		&testMetaRepo{deleteErr: errors.New("database unavailable")},
		&testSessionTracker{},
		&testProvider{},
	)

	err := srv.Delete(context.Background(), "local.test://file.bin")

	require.ErrorIs(t, err, storageapi.ErrFailToDelete)
}

func TestInitMultipartStoreKeepsSessionWhenRollbackFails(t *testing.T) {
	trackerDeleted := false
	tracker := &testSessionTracker{
		saveErr: errors.New("redis unavailable"),
		deleteFn: func(context.Context, string) error {
			trackerDeleted = true
			return nil
		},
	}
	provider := &testProvider{abortFn: func(context.Context, string) error {
		return errors.New("s3 abort failed")
	}}
	srv := newTestService(
		&testMetaRepo{loadErr: storageapi.ErrMetaNotFound},
		tracker,
		provider,
	)

	_, _, err := srv.InitMultipartStore(context.Background(), "local.test", "file.bin", storageapi.FileMeta{})

	require.ErrorIs(t, err, storageapi.ErrInitPartUploadFailed)
	require.False(t, trackerDeleted)
}

func TestCompleteMultipartDoesNotReportSuccessWhenMetadataFinalizationFails(t *testing.T) {
	storageKey := storageapi.StorageKey("local.test://file.bin")
	sessionDeleted := false
	tracker := &testSessionTracker{
		getFn: func(context.Context, string) (storageapi.UploadSession, error) {
			return storageapi.UploadSession{StorageKey: storageKey}, nil
		},
		deleteFn: func(context.Context, string) error {
			sessionDeleted = true
			return nil
		},
	}
	srv := newTestService(
		&testMetaRepo{
			loadFn: func(context.Context, storageapi.StorageKey) (storageapi.FileMeta, error) {
				return storageapi.FileMeta{StorageKey: storageKey}, nil
			},
			storeErr: errors.New("database unavailable"),
		},
		tracker,
		&testProvider{},
	)

	_, err := srv.CompleteMultipart(context.Background(), "upload")

	require.ErrorIs(t, err, storageapi.ErrStorePartFailed)
	require.False(t, sessionDeleted)
}

func TestCompleteMultipartFinalizesWithDetachedContext(t *testing.T) {
	type contextKey struct{}
	storageKey := storageapi.StorageKey("local.test://file.bin")
	requestCtx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request-value"))
	provider := &testProvider{completeFn: func(context.Context, string) error {
		cancel()
		return nil
	}}
	tracker := &testSessionTracker{
		getFn: func(context.Context, string) (storageapi.UploadSession, error) {
			return storageapi.UploadSession{StorageKey: storageKey}, nil
		},
		deleteFn: func(ctx context.Context, _ string) error {
			require.NoError(t, ctx.Err())
			require.Equal(t, "request-value", ctx.Value(contextKey{}))
			return nil
		},
	}
	srv := newTestService(
		&testMetaRepo{
			loadFn: func(ctx context.Context, _ storageapi.StorageKey) (storageapi.FileMeta, error) {
				require.NoError(t, ctx.Err())
				return storageapi.FileMeta{StorageKey: storageKey}, nil
			},
			storeFn: func(ctx context.Context, meta storageapi.FileMeta) error {
				require.NoError(t, ctx.Err())
				require.True(t, meta.Finished)
				return nil
			},
		},
		tracker,
		provider,
	)

	meta, err := srv.CompleteMultipart(requestCtx, "upload")

	require.NoError(t, err)
	require.Equal(t, storageKey, meta.StorageKey)
	require.True(t, meta.Finished)
}

func TestAbortMultipartReturnsErrorWhenSessionCleanupFails(t *testing.T) {
	tracker := &testSessionTracker{
		getFn: func(context.Context, string) (storageapi.UploadSession, error) {
			return storageapi.UploadSession{StorageKey: "local.test://file.bin"}, nil
		},
		deleteErr: errors.New("redis unavailable"),
	}
	srv := newTestService(&testMetaRepo{}, tracker, &testProvider{})

	err := srv.AbortMultipart(context.Background(), "upload")

	require.ErrorIs(t, err, storageapi.ErrFailToAbortMultipartStore)
}

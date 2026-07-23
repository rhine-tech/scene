package sessiontracker

import (
	"context"
	"testing"
	"time"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
)

type testRedisDataSource struct {
	datasource.RedisDataSource
	contexts []context.Context
	value    string
}

func (d *testRedisDataSource) Set(ctx context.Context, _ string, value interface{}, _ time.Duration) error {
	d.contexts = append(d.contexts, ctx)
	d.value = value.(string)
	return nil
}

func (d *testRedisDataSource) Get(ctx context.Context, _ string) (string, error) {
	d.contexts = append(d.contexts, ctx)
	return d.value, nil
}

func (d *testRedisDataSource) Delete(ctx context.Context, _ string) error {
	d.contexts = append(d.contexts, ctx)
	return nil
}

func TestRedisUploadSessionTrackerPassesContext(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "request-value")
	dataSource := &testRedisDataSource{}
	tracker := NewRedisUploadSessionTracker().(*redisUploadSessionTracker)
	tracker.redisDs = dataSource
	session := storage.UploadSession{StorageKey: "s3.default://file.bin"}

	require.NoError(t, tracker.Save(ctx, "upload", session))
	loaded, err := tracker.Get(ctx, "upload")
	require.NoError(t, err)
	require.Equal(t, session, loaded)
	require.NoError(t, tracker.Delete(ctx, "upload"))
	require.Len(t, dataSource.contexts, 3)
	for _, gotCtx := range dataSource.contexts {
		require.Equal(t, "request-value", gotCtx.Value(contextKey{}))
	}
}

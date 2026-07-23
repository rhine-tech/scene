package meta

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/rhine-tech/scene"
	sceneorm "github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/infrastructure/logger"
	loggerrepo "github.com/rhine-tech/scene/infrastructure/logger/repository"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

var registerRepositoryTestLogger sync.Once

func newRepositoryTestGorm(t *testing.T) *sceneorm.Gorm {
	t.Helper()

	registerRepositoryTestLogger.Do(func() {
		registry.Register[logger.ILogger](loggerrepo.NewZapColoredLogger())
	})

	dsnName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	ds := datasources.SqliteDatasource(datasource.SqliteConfig{
		Path:    "file:" + dsnName,
		Options: "mode=memory&cache=shared",
	})
	registry.Inject(ds)
	require.NoError(t, ds.Setup())
	ds.Connection().SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, ds.Dispose())
	})

	db := sceneorm.NewGormWithSQLite(ds)
	registry.Inject(db)
	require.NoError(t, db.Setup())
	return db
}

func TestGormFileMetaRepositoryUpsert(t *testing.T) {
	ctx := context.Background()
	repo := NewGormFileMetaRepository(newRepositoryTestGorm(t))
	require.NoError(t, repo.(scene.Setupable).Setup())

	meta := storage.FileMeta{
		StorageKey:       "local.test://covers/work.jpg",
		Provider:         "local.test",
		Identifier:       "covers/work.jpg",
		OriginalFilename: "old.jpg",
		ContentType:      "image/jpeg",
		ContentLength:    10,
		Md5Checksum:      "old",
		Finished:         false,
	}
	require.NoError(t, repo.Store(ctx, meta))

	meta.OriginalFilename = "new.jpg"
	meta.ContentLength = 20
	meta.Md5Checksum = "new"
	meta.Finished = true
	require.NoError(t, repo.Store(ctx, meta))

	loaded, err := repo.Load(ctx, meta.StorageKey)
	require.NoError(t, err)
	require.Equal(t, meta.StorageKey, loaded.StorageKey)
	require.Equal(t, meta.Provider, loaded.Provider)
	require.Equal(t, meta.Identifier, loaded.Identifier)
	require.Equal(t, meta.OriginalFilename, loaded.OriginalFilename)
	require.Equal(t, meta.ContentLength, loaded.ContentLength)
	require.Equal(t, meta.Md5Checksum, loaded.Md5Checksum)
	require.True(t, loaded.Finished)

	page, err := repo.List(ctx, meta.Provider, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(1), page.Count)

	require.NoError(t, repo.Delete(ctx, meta.StorageKey))
	_, err = repo.Load(ctx, meta.StorageKey)
	require.ErrorIs(t, err, storage.ErrMetaNotFound)
}

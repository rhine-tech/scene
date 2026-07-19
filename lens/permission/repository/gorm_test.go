package repository

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
	ds := datasources.SqliteDatasource(datasource.DatabaseConfig{
		Host:    "file:" + dsnName,
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

func TestGormPermissionRepositoryAddIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := NewGormImpl(newRepositoryTestGorm(t))
	require.NoError(t, repo.(scene.Setupable).Setup())

	for range 2 {
		_, err := repo.AddPermission(ctx, "user-1", "notify:send")
		require.NoError(t, err)
	}

	permissions, err := repo.GetPermissions(ctx, "user-1")
	require.NoError(t, err)
	require.Len(t, permissions, 1)
	require.Equal(t, "notify:send", permissions[0].String())

	require.NoError(t, repo.RemovePermission(ctx, "user-1", "notify:send"))
	permissions, err = repo.GetPermissions(ctx, "user-1")
	require.NoError(t, err)
	require.Empty(t, permissions)
}

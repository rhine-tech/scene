package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/rhine-tech/scene"
	sceneorm "github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
)

func newRepositoryTestGorm(t *testing.T) *sceneorm.Gorm {
	t.Helper()

	container := registry.NewContainer()
	registry.Export[logger.ILogger](container, logger.NoopLogger{})

	dsnName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	ds := datasources.SqliteDatasource(datasource.SqliteConfig{
		Path:    "file:" + dsnName,
		Options: "mode=memory&cache=shared",
	})
	db := sceneorm.NewGormWithSQLite(ds)
	container.Load(ds)
	container.Load(db)
	scope := registry.NewScope()
	require.NoError(t, scope.Build(container))
	require.NoError(t, ds.Setup())
	ds.Connection().SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, db.TearDown())
		require.NoError(t, ds.TearDown())
	})

	require.NoError(t, db.Setup())
	return db
}

func TestGormPermissionRepositoryAddIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := NewGormImpl(newRepositoryTestGorm(t))
	require.NoError(t, repo.(scene.Lifecycle).Setup())

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

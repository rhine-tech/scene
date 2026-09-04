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
	"github.com/rhine-tech/scene/lens/permission"
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

func newPermissionRepository(t *testing.T) permission.PermissionRepository {
	t.Helper()
	repo := NewGormImpl(newRepositoryTestGorm(t))
	require.NoError(t, repo.(scene.Lifecycle).Setup())
	return repo
}

func permissionStrings(permissions []*permission.Permission) []string {
	values := make([]string, len(permissions))
	for i, value := range permissions {
		values[i] = value.String()
	}
	return values
}

func TestGormPermissionRepositoryAddIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := newPermissionRepository(t)

	perm := permission.MustParsePermission("notify:send")
	for range 2 {
		require.NoError(t, repo.AddPermission(ctx, "user-1", perm))
	}

	permissions, err := repo.GetPermissions(ctx, "user-1")
	require.NoError(t, err)
	require.Len(t, permissions, 1)
	require.Equal(t, "notify:send", permissions[0].String())

	require.NoError(t, repo.RemovePermission(ctx, "user-1", perm))
	permissions, err = repo.GetPermissions(ctx, "user-1")
	require.NoError(t, err)
	require.Empty(t, permissions)
}

func TestGormPermissionRepositoryHasPermission(t *testing.T) {
	ctx := context.Background()
	repo := newPermissionRepository(t)
	for owner, value := range map[string]string{
		"parent":  "project",
		"branch":  "project:123",
		"exact":   "project:123:read",
		"child":   "project:123:read:detail",
		"sibling": "project:456",
	} {
		require.NoError(t, repo.AddPermission(ctx, owner, permission.MustParsePermission(value)))
	}

	requested := permission.MustParsePermission("project:123:read")
	for _, owner := range []string{"parent", "branch", "exact"} {
		allowed, err := repo.HasPermission(ctx, owner, requested)
		require.NoError(t, err)
		require.True(t, allowed, owner)
	}
	for _, owner := range []string{"child", "sibling", "missing"} {
		allowed, err := repo.HasPermission(ctx, owner, requested)
		require.NoError(t, err)
		require.False(t, allowed, owner)
	}
}

func TestGormPermissionRepositoryReplacePermissions(t *testing.T) {
	ctx := context.Background()
	repo := newPermissionRepository(t)
	for _, value := range []string{
		"project",
		"project:123",
		"project:123:read",
		"project:123:document:1:read",
		"project:456:read",
	} {
		require.NoError(t, repo.AddPermission(ctx, "owner", permission.MustParsePermission(value)))
	}

	prefix := permission.MustParsePermission("project:123")
	replacement := permission.MustParsePermission("project:123:write")
	require.NoError(t, repo.ReplacePermissions(ctx, "owner", prefix, []*permission.Permission{
		replacement,
		replacement.Copy(),
	}))
	permissions, err := repo.GetPermissions(ctx, "owner")
	require.NoError(t, err)
	require.Equal(t, []string{"project", "project:123:write", "project:456:read"}, permissionStrings(permissions))

	err = repo.ReplacePermissions(ctx, "owner", prefix, []*permission.Permission{
		permission.MustParsePermission("project:123:admin"),
		permission.MustParsePermission("project:456:admin"),
	})
	require.Error(t, err)
	permissions, err = repo.GetPermissions(ctx, "owner")
	require.NoError(t, err)
	require.Equal(t, []string{"project", "project:123:write", "project:456:read"}, permissionStrings(permissions))

	require.NoError(t, repo.ReplacePermissions(ctx, "owner", prefix, nil))
	permissions, err = repo.GetPermissions(ctx, "owner")
	require.NoError(t, err)
	require.Equal(t, []string{"project", "project:456:read"}, permissionStrings(permissions))
}

func TestGormPermissionRepositoryListOwnersWithPermission(t *testing.T) {
	ctx := context.Background()
	repo := newPermissionRepository(t)
	grants := map[string][]string{
		"alice": {"project", "project:123"},
		"bob":   {"project:123:read"},
		"carol": {"project:123:read:detail"},
		"dave":  {"project:456"},
	}
	for owner, values := range grants {
		for _, value := range values {
			require.NoError(t, repo.AddPermission(ctx, owner, permission.MustParsePermission(value)))
		}
	}

	requested := permission.MustParsePermission("project:123:read")
	result, err := repo.ListOwnersWithPermission(ctx, requested, 0, 20)
	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
	require.Equal(t, int64(2), result.Count)
	require.Equal(t, int64(0), result.Offset)
	require.Equal(t, []permission.OwnerPermissions{
		{Owner: "alice", Permissions: []*permission.Permission{
			permission.MustParsePermission("project"),
			permission.MustParsePermission("project:123"),
		}},
		{Owner: "bob", Permissions: []*permission.Permission{
			permission.MustParsePermission("project:123:read"),
		}},
	}, result.Results)

	page, err := repo.ListOwnersWithPermission(ctx, requested, 1, 1)
	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Equal(t, int64(1), page.Count)
	require.Equal(t, int64(1), page.Offset)
	require.Equal(t, "bob", page.Results[0].Owner)
}

func TestGormPermissionRepositoryListOwnersWithGrantsByPrefix(t *testing.T) {
	ctx := context.Background()
	repo := newPermissionRepository(t)
	grants := map[string][]string{
		"alice": {"project"},
		"bob":   {"project:123", "project:123:read"},
		"carol": {"project:123:file:1:read"},
		"dave":  {"project:1234:read"},
		"eve":   {"project:456"},
	}
	for owner, values := range grants {
		for _, value := range values {
			require.NoError(t, repo.AddPermission(ctx, owner, permission.MustParsePermission(value)))
		}
	}

	result, err := repo.ListOwnersWithGrantsByPrefix(ctx, permission.MustParsePermission("project:123"), 0, 20)
	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
	require.Equal(t, int64(2), result.Count)
	require.Equal(t, []permission.OwnerPermissions{
		{Owner: "bob", Permissions: []*permission.Permission{
			permission.MustParsePermission("project:123"),
			permission.MustParsePermission("project:123:read"),
		}},
		{Owner: "carol", Permissions: []*permission.Permission{
			permission.MustParsePermission("project:123:file:1:read"),
		}},
	}, result.Results)

	page, err := repo.ListOwnersWithGrantsByPrefix(ctx, permission.MustParsePermission("project:123"), 1, 1)
	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Equal(t, int64(1), page.Count)
	require.Equal(t, int64(1), page.Offset)
	require.Equal(t, "carol", page.Results[0].Owner)
}

func TestGormPermissionRepositoryRemovePermissionsByPrefix(t *testing.T) {
	ctx := context.Background()
	repo := newPermissionRepository(t)
	grants := map[string][]string{
		"alice": {"project:a_b"},
		"bob":   {"project:a_b:read"},
		"carol": {"project:axb:read"},
		"dave":  {"project", "project:other"},
	}
	for owner, values := range grants {
		for _, value := range values {
			require.NoError(t, repo.AddPermission(ctx, owner, permission.MustParsePermission(value)))
		}
	}

	require.NoError(t, repo.RemovePermissionsByPrefix(ctx, permission.MustParsePermission("project:a_b")))

	for _, owner := range []string{"alice", "bob"} {
		permissions, err := repo.GetPermissions(ctx, owner)
		require.NoError(t, err)
		require.Empty(t, permissions)
	}
	permissions, err := repo.GetPermissions(ctx, "carol")
	require.NoError(t, err)
	require.Equal(t, []string{"project:axb:read"}, permissionStrings(permissions))
	permissions, err = repo.GetPermissions(ctx, "dave")
	require.NoError(t, err)
	require.Equal(t, []string{"project", "project:other"}, permissionStrings(permissions))
}

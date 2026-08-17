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
	"github.com/rhine-tech/scene/lens/authentication"
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

func TestGormAuthenticationRepositories(t *testing.T) {
	ctx := context.Background()
	db := newRepositoryTestGorm(t)

	users := NewGormAuthenticationRepository(db)
	require.NoError(t, users.(scene.Lifecycle).Setup())

	alice := authentication.User{
		UserID:   "user-1",
		Username: "alice",
		Password: "secret",
		Email:    "alice@example.com",
	}
	created, err := users.AddUser(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, alice, created)

	loaded, err := users.UserById(ctx, alice.UserID)
	require.NoError(t, err)
	require.Equal(t, alice, loaded)

	userID, err := users.Authenticate(ctx, alice.Username, alice.Password)
	require.NoError(t, err)
	require.Equal(t, alice.UserID, userID)

	_, err = users.AddUser(ctx, authentication.User{
		UserID:   "user-2",
		Username: alice.Username,
	})
	require.ErrorIs(t, err, authentication.ErrUserAlreadyExists)

	alice.DisplayName = ""
	alice.Email = "renamed@example.com"
	require.NoError(t, users.UpdateUser(ctx, alice))
	updated, err := users.UserByEmail(ctx, alice.Email)
	require.NoError(t, err)
	require.Equal(t, alice, updated)

	page, err := users.ListUsers(ctx, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(1), page.Count)
	require.Equal(t, []authentication.User{alice}, page.Results)

	tokens := NewGormAccessTokenRepository(db)
	require.NoError(t, tokens.(scene.Lifecycle).Setup())

	for _, token := range []authentication.AccessToken{
		{Token: "token-1", UserID: alice.UserID, Name: "first"},
		{Token: "token-2", UserID: alice.UserID, Name: "second"},
		{Token: "token-3", UserID: "user-else", Name: "other"},
	} {
		_, err := tokens.CreateToken(ctx, token)
		require.NoError(t, err)
	}

	tokenPage, err := tokens.ListTokensByUser(ctx, alice.UserID, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), tokenPage.Total)
	require.Equal(t, int64(2), tokenPage.Count)
}

package base

import (
	"context"
	"errors"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	"github.com/stretchr/testify/require"
)

type recordingPermissionRepository struct {
	err         error
	contexts    []context.Context
	permissions []*permission.Permission
}

func (r *recordingPermissionRepository) ImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionRepository", "recording")
}

func (r *recordingPermissionRepository) GetPermissions(ctx context.Context, _ string) ([]*permission.Permission, error) {
	r.contexts = append(r.contexts, ctx)
	return nil, r.err
}

func (r *recordingPermissionRepository) HasPermission(ctx context.Context, _ string, _ *permission.Permission) (bool, error) {
	r.contexts = append(r.contexts, ctx)
	return false, r.err
}

func (r *recordingPermissionRepository) AddPermission(ctx context.Context, _ string, perm *permission.Permission) error {
	r.contexts = append(r.contexts, ctx)
	r.permissions = append(r.permissions, perm)
	return r.err
}

func (r *recordingPermissionRepository) RemovePermission(ctx context.Context, _ string, perm *permission.Permission) error {
	r.contexts = append(r.contexts, ctx)
	r.permissions = append(r.permissions, perm)
	return r.err
}

func (r *recordingPermissionRepository) ReplacePermissions(ctx context.Context, _ string, _ *permission.Permission, _ []*permission.Permission) error {
	r.contexts = append(r.contexts, ctx)
	return r.err
}

func (r *recordingPermissionRepository) ListOwnersWithPermission(ctx context.Context, _ *permission.Permission, _, _ int64) (model.PaginationResult[permission.OwnerPermissions], error) {
	r.contexts = append(r.contexts, ctx)
	return model.PaginationResult[permission.OwnerPermissions]{}, r.err
}

func (r *recordingPermissionRepository) ListOwnersWithGrantsByPrefix(ctx context.Context, _ *permission.Permission, _, _ int64) (model.PaginationResult[permission.OwnerPermissions], error) {
	r.contexts = append(r.contexts, ctx)
	return model.PaginationResult[permission.OwnerPermissions]{}, r.err
}

func (r *recordingPermissionRepository) RemovePermissionsByPrefix(ctx context.Context, _ *permission.Permission) error {
	r.contexts = append(r.contexts, ctx)
	return r.err
}

type serviceContextKey struct{}

func TestPermissionManagerPropagatesContextAndRepositoryErrors(t *testing.T) {
	repositoryErr := errors.New("repository unavailable")
	repository := &recordingPermissionRepository{err: repositoryErr}
	service := &PermissionManagerImpl{
		logger: logger.NoopLogger{},
		repo:   repository,
	}
	ctx := context.WithValue(context.Background(), serviceContextKey{}, "request")

	_, err := service.HasPermission(ctx, "owner", permission.MustParsePermission("project:read"))
	require.ErrorIs(t, err, repositoryErr)
	_, err = service.ListPermissions(ctx, "owner")
	require.ErrorIs(t, err, repositoryErr)
	require.ErrorIs(t, service.AddPermission(ctx, "owner", "project:read"), repositoryErr)
	require.ErrorIs(t, service.RemovePermission(ctx, "owner", "project:read"), repositoryErr)
	require.ErrorIs(t, service.ReplacePermissions(ctx, "owner", permission.MustParsePermission("project"), nil), repositoryErr)
	_, err = service.ListOwnersWithPermission(ctx, permission.MustParsePermission("project:read"), 0, 20)
	require.ErrorIs(t, err, repositoryErr)
	_, err = service.ListOwnersWithGrantsByPrefix(ctx, permission.MustParsePermission("project"), 0, 20)
	require.ErrorIs(t, err, repositoryErr)
	require.ErrorIs(t, service.RemovePermissionsByPrefix(ctx, permission.MustParsePermission("project")), repositoryErr)

	require.Len(t, repository.contexts, 8)
	for _, repositoryCtx := range repository.contexts {
		require.Equal(t, "request", repositoryCtx.Value(serviceContextKey{}))
	}
}

func TestPermissionManagerParsesMutationPermissionsBeforeRepository(t *testing.T) {
	repository := &recordingPermissionRepository{}
	service := &PermissionManagerImpl{
		logger: logger.NoopLogger{},
		repo:   repository,
	}
	ctx := context.Background()

	require.NoError(t, service.AddPermission(ctx, "owner", "project:read"))
	require.NoError(t, service.RemovePermission(ctx, "owner", "project:write"))
	require.Equal(t, "project:read", repository.permissions[0].String())
	require.Equal(t, "project:write", repository.permissions[1].String())

	require.Error(t, service.AddPermission(ctx, "owner", "project::read"))
	require.Error(t, service.RemovePermission(ctx, "owner", "project::write"))
	require.Len(t, repository.permissions, 2)
}

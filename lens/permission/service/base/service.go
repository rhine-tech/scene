package base

import (
	"context"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
)

type PermissionManagerImpl struct {
	logger logger.ILogger                  `aperture:""`
	repo   permission.PermissionRepository `aperture:""`
}

// NewPermissionManager creates a permission service backed by repo.
func NewPermissionManager(repo permission.PermissionRepository) permission.PermissionService {
	return &PermissionManagerImpl{repo: repo}
}

func (p *PermissionManagerImpl) ImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionService", "default")
}

func (p *PermissionManagerImpl) Setup() error {
	p.logger = p.logger.WithPrefix(p.ImplName().Identifier())
	p.logger.Infof("Permission service is ready, using: %s", p.repo.ImplName())
	return nil
}

func (p *PermissionManagerImpl) TearDown() error {
	return nil
}

func (p *PermissionManagerImpl) HasPermission(ctx context.Context, owner string, perm *permission.Permission) (bool, error) {
	allowed, err := p.repo.HasPermission(ctx, owner, perm)
	if err != nil {
		p.logger.ErrorW("failed to check permission", "owner", owner, "permission", perm, "error", err)
		return false, err
	}
	return allowed, nil
}

func (p *PermissionManagerImpl) HasPermissionStr(ctx context.Context, owner string, perm string) (bool, error) {
	p1, err := permission.ParsePermission(perm)
	if err != nil {
		return false, err
	}
	return p.HasPermission(ctx, owner, p1)
}

func (p *PermissionManagerImpl) ListPermissions(ctx context.Context, role string) ([]*permission.Permission, error) {
	perms, err := p.repo.GetPermissions(ctx, role)
	if err != nil {
		p.logger.ErrorW("failed to list permissions", "owner", role, "error", err)
		return nil, err
	}
	return perms, nil
}

func (p *PermissionManagerImpl) AddPermission(ctx context.Context, role string, perm string) error {
	parsed, err := permission.ParsePermission(perm)
	if err != nil {
		p.logger.Errorf("failed to add permission %s: %s", perm, err)
		return err
	}
	if err := p.repo.AddPermission(ctx, role, parsed); err != nil {
		p.logger.Errorf("failed to add permission %s: %s", perm, err)
		return err
	}
	return nil
}

func (p *PermissionManagerImpl) RemovePermission(ctx context.Context, role string, perm string) error {
	parsed, err := permission.ParsePermission(perm)
	if err != nil {
		p.logger.Errorf("failed to remove permission %s: %s", perm, err)
		return err
	}
	if err := p.repo.RemovePermission(ctx, role, parsed); err != nil {
		p.logger.Errorf("failed to remove permission %s: %s", perm, err)
		return err
	}
	return nil
}

func (p *PermissionManagerImpl) ReplacePermissions(
	ctx context.Context,
	owner string,
	prefix *permission.Permission,
	permissions []*permission.Permission,
) error {
	if err := p.repo.ReplacePermissions(ctx, owner, prefix, permissions); err != nil {
		p.logger.ErrorW("failed to replace permissions", "owner", owner, "prefix", prefix, "error", err)
		return err
	}
	return nil
}

func (p *PermissionManagerImpl) ListOwnersWithPermission(
	ctx context.Context,
	perm *permission.Permission,
	offset, limit int64,
) (model.PaginationResult[permission.OwnerPermissions], error) {
	result, err := p.repo.ListOwnersWithPermission(ctx, perm, offset, limit)
	if err != nil {
		p.logger.ErrorW("failed to list owners with permission", "permission", perm, "error", err)
		return result, err
	}
	return result, nil
}

func (p *PermissionManagerImpl) ListOwnersWithGrantsByPrefix(
	ctx context.Context,
	prefix *permission.Permission,
	offset, limit int64,
) (model.PaginationResult[permission.OwnerPermissions], error) {
	result, err := p.repo.ListOwnersWithGrantsByPrefix(ctx, prefix, offset, limit)
	if err != nil {
		p.logger.ErrorW("failed to list owners by permission prefix", "prefix", prefix, "error", err)
		return result, err
	}
	return result, nil
}

func (p *PermissionManagerImpl) RemovePermissionsByPrefix(
	ctx context.Context,
	prefix *permission.Permission,
) error {
	if err := p.repo.RemovePermissionsByPrefix(ctx, prefix); err != nil {
		p.logger.ErrorW("failed to remove permissions by prefix", "prefix", prefix, "error", err)
		return err
	}
	return nil
}

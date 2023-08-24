package service

import (
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/lens/middlewares/permission"
)

type permissionManager struct {
	logger logger.ILogger
	repo   permission.PermissionRepository
}

func (p *permissionManager) SrvImplName() string {
	return "permission.service.manage"
}

func NewPermissionManager(logger logger.ILogger, repo permission.PermissionRepository) permission.PermissionService {
	pm := &permissionManager{
		repo: repo,
	}
	pm.logger = logger.WithPrefix(pm.SrvImplName())
	if err := pm.repo.Status(); err != nil {
		pm.logger.Errorf("Failed to reload permission repository: %s", err.Error())
	}
	return pm
}

func (p *permissionManager) HasPermission(owner string, perm string) bool {
	p1, err := permission.ParsePermission(perm)
	if err != nil {
		return false
	}
	perms := p.repo.GetPermissions(owner)
	for _, p0 := range perms {
		if p0.HasPermission(p1) {
			return true
		}
	}
	return false
}

//func (p *permissionManager) ListOwners() []string {
//	owners := p.repo.GetOwners()
//	names := make([]string, len(owners))
//	for i, owner := range owners {
//		names[i] = string(owner)
//	}
//	return names
//}

func (p *permissionManager) ListPermissions(role string) permission.PermissionSet {
	return p.repo.GetPermissions(role)
}

func (p *permissionManager) AddPermission(role string, perm string) error {
	_, err := p.repo.AddPermission(role, perm)
	if err != nil {
		p.logger.Errorf("failed to add permission %s: %s", perm, err)
		return err
	}
	return nil
}

func (p *permissionManager) RemovePermission(role string, perm string) error {
	err := p.repo.RemovePermission(role, perm)
	if err != nil {
		p.logger.Errorf("failed to remove permission %s: %s", perm, err)
		return err
	}
	return nil
}

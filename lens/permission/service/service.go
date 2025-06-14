package service

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/permission"
)

type PermissionManagerImpl struct {
	logger logger.ILogger                  `aperture:""`
	repo   permission.PermissionRepository `aperture:""`
}

func (p *PermissionManagerImpl) SrvImplName() scene.ImplName {
	return permission.Lens.ImplName("PermissionManager", "default")
}

func (p *PermissionManagerImpl) Setup() error {
	p.logger = p.logger.WithPrefix(p.SrvImplName().Identifier())
	//if err := p.repo.Status(); err != nil {
	//	p.logger.Errorf("Failed to reload permission repository: %s", err.Error())
	//}
	p.logger.Infof("Permission service is ready, using: %s", p.repo.ImplName())
	return nil
}

func (p *PermissionManagerImpl) HasPermission(owner string, perm *permission.Permission) bool {
	perms := p.repo.GetPermissions(owner)
	for _, p0 := range perms {
		if p0.HasPermission(perm) {
			return true
		}
	}
	return false
}

func (p *PermissionManagerImpl) HasPermissionStr(owner string, perm string) bool {
	p1, err := permission.ParsePermission(perm)
	if err != nil {
		return false
	}
	return p.HasPermission(owner, p1)
}

//func (p *PermissionManagerImpl) ListOwners() []string {
//	owners := p.repo.GetOwners()
//	names := make([]string, len(owners))
//	for i, owner := range owners {
//		names[i] = string(owner)
//	}
//	return names
//}

func (p *PermissionManagerImpl) ListPermissions(role string) permission.PermissionSet {
	return p.repo.GetPermissions(role)
}

func (p *PermissionManagerImpl) AddPermission(role string, perm string) error {
	_, err := p.repo.AddPermission(role, perm)
	if err != nil {
		p.logger.Errorf("failed to add permission %s: %s", perm, err)
		return err
	}
	return nil
}

func (p *PermissionManagerImpl) RemovePermission(role string, perm string) error {
	err := p.repo.RemovePermission(role, perm)
	if err != nil {
		p.logger.Errorf("failed to remove permission %s: %s", perm, err)
		return err
	}
	return nil
}

package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/repository"
	"github.com/rhine-tech/scene/lens/permission/service/base"
	"github.com/rhine-tech/scene/registry"
)

type ServiceGorm struct {
	scene.ModuleFactory
}

func (b ServiceGorm) Init(container *registry.Container) {
	repositoryImpl := registry.Load(container, repository.NewGormImpl(nil))
	registry.Export[permission.PermissionService](container, base.NewPermissionManager(repositoryImpl))
}

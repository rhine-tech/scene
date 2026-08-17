package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/gen/arpcimpl"
	"github.com/rhine-tech/scene/lens/permission/repository"
	"github.com/rhine-tech/scene/lens/permission/service/base"
	"github.com/rhine-tech/scene/lens/permission/service/proxy"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
)

type ServiceARpc struct {
	scene.ModuleFactory
	Client sarpc.Client
}

func (b ServiceARpc) Init(container *registry.Container) {
	registry.Export[permission.PermissionService](container, arpcimpl.NewARpcPermissionService(b.Client))
}

type ServiceGorm struct {
	scene.ModuleFactory
}

func (b ServiceGorm) Init(container *registry.Container) {
	repositoryImpl := registry.Load(container, repository.NewGormImpl(nil))
	baseService := registry.Load(container, base.NewPermissionManager(repositoryImpl))
	registry.Export[permission.PermissionService](container, proxy.NewCachedPermissionService(baseService))
}

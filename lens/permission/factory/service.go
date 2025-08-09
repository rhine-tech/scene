package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/gen/arpcimpl"
	"github.com/rhine-tech/scene/lens/permission/repository"
	"github.com/rhine-tech/scene/lens/permission/service"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
)

type ServiceARpc struct {
	scene.ModuleFactory
	Client sarpc.Client
}

func (b ServiceARpc) Init() scene.LensInit {
	return func() {
		registry.Register[permission.PermissionService](arpcimpl.NewARpcPermissionService(b.Client))
	}
}

type ServiceGorm struct {
	scene.ModuleFactory
}

func (b ServiceGorm) Init() scene.LensInit {
	return func() {
		_ = registry.Register(repository.NewGormImpl(nil))
		_ = registry.Register(
			permission.PermissionService(&service.PermissionManagerImpl{}))
	}
}

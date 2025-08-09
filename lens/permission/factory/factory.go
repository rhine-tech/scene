package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/delivery"
	"github.com/rhine-tech/scene/lens/permission/gen/arpcimpl"
	"github.com/rhine-tech/scene/lens/permission/repository"
	"github.com/rhine-tech/scene/lens/permission/service"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func InitApp() sgin.GinApplication {
	return delivery.NewGinApp(registry.AcquireSingleton(permission.PermissionService(nil)))
}

type Gorm struct {
	scene.ModuleFactory
}

func (b Gorm) Init() scene.LensInit {
	return func() {
		_ = registry.Register(repository.NewGormImpl(nil))
		_ = registry.Register(
			permission.PermissionService(&service.PermissionManagerImpl{}))
	}
}

func (b Gorm) Apps() []any {
	return []any{
		InitApp,
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](new(arpcimpl.ARpcAppPermissionService))
		},
	}
}

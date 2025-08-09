package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/delivery"
	"github.com/rhine-tech/scene/lens/permission/gen/arpcimpl"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func InitApp() sgin.GinApplication {
	return delivery.NewGinApp(registry.AcquireSingleton(permission.PermissionService(nil)))
}

type AppGin struct {
	scene.ModuleFactory
}

func (b AppGin) Apps() []any {
	return []any{
		InitApp,
	}
}

type AppARpc struct {
	scene.ModuleFactory
}

func (b AppARpc) Apps() []any {
	return []any{
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](new(arpcimpl.ARpcAppPermissionService))
		},
	}
}

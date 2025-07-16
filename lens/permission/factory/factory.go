package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/delivery"
	"github.com/rhine-tech/scene/lens/permission/repository"
	"github.com/rhine-tech/scene/lens/permission/service"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func InitApp() sgin.GinApplication {
	return delivery.NewGinApp(registry.AcquireSingleton(permission.PermissionService(nil)))
}

//type MongoCached struct {
//	scene.ModuleFactory
//}
//
//func (b MongoCached) Init() scene.LensInit {
//	return func() {
//		_ = registry.Register(repository.NewPermissionMongoRepoCached())
//		_ = registry.Register(
//			permission.PermissionService(&service.PermissionManagerImpl{}))
//	}
//}
//
//func (b MongoCached) Apps() []any {
//	return []any{
//		InitApp,
//	}
//}

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
	}
}

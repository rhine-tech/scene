package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/lens/permission/delivery"
	"github.com/rhine-tech/scene/lens/permission/repository"
	"github.com/rhine-tech/scene/lens/permission/service"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

// Init is instance of scene.LensInit
func Init() {
	cfg := registry.AcquireSingleton(&model.DatabaseConfig{})
	_ = registry.Register(repository.NewPermissionMongoRepo(*cfg))
	_ = registry.Register(
		permission.PermissionService(&service.PermissionManagerImpl{}))
	return
}

func InitApp() sgin.GinApplication {
	return delivery.NewGinApp(registry.AcquireSingleton(logger.ILogger(nil)), registry.AcquireSingleton(permission.PermissionService(nil)))
}

// Deprecated: use MongoNoCache instead
type MongoNoCache struct {
	scene.ModuleFactory
}

func (b MongoNoCache) Init() scene.LensInit {
	return Init
}

func (b MongoNoCache) Apps() []any {
	return []any{
		InitApp,
	}
}

type MongoCached struct {
	scene.ModuleFactory
}

func (b MongoCached) Init() scene.LensInit {
	return func() {
		_ = registry.Register(repository.NewPermissionMongoRepoCached())
		_ = registry.Register(
			permission.PermissionService(&service.PermissionManagerImpl{}))
	}
}

func (b MongoCached) Apps() []any {
	return []any{
		InitApp,
	}
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
	}
}

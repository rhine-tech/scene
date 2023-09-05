package builder

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/permission"
	"github.com/rhine-tech/scene/lens/middlewares/permission/delivery"
	"github.com/rhine-tech/scene/lens/middlewares/permission/repository"
	"github.com/rhine-tech/scene/lens/middlewares/permission/service"
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

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

func (b Builder) Apps() []any {
	return []any{
		InitApp,
	}
}

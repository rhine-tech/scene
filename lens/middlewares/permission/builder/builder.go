package builder

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/lens/middlewares/permission"
	"github.com/aynakeya/scene/lens/middlewares/permission/delivery"
	"github.com/aynakeya/scene/lens/middlewares/permission/repository"
	"github.com/aynakeya/scene/lens/middlewares/permission/service"
	"github.com/aynakeya/scene/model"
	"github.com/aynakeya/scene/registry"
	sgin "github.com/aynakeya/scene/scenes/gin"
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

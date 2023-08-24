package builder

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/datasource/repository"
	"github.com/aynakeya/scene/model"
	"github.com/aynakeya/scene/registry"
)

// Init is instance of scene.LensInit
func Init() {
	cfg := registry.Register(&model.DatabaseConfig{
		Host:     registry.Config.GetString("scene.db.host"),
		Port:     int(registry.Config.GetInt("scene.db.port")),
		Username: registry.Config.GetString("scene.db.username"),
		Password: registry.Config.GetString("scene.db.password"),
		Database: "scene",
	})
	registry.Register(
		repository.NewMongoDataSource(*cfg))

}

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

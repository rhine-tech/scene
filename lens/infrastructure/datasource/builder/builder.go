package builder

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource/repository"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
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
	registry.Register(
		repository.NewRedisDataRepo(
			model.DatabaseConfig{
				Host:     registry.Config.GetString("scene.redis.host"),
				Port:     int(registry.Config.GetInt("scene.redis.port")),
				Database: "0",
			}))

}

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

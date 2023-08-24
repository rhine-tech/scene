package builder

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/config"
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/lens/infrastructure/logger/repository"
	"github.com/aynakeya/scene/registry"
)

// Init is instance of scene.LensInit
func Init() {
	cfg := registry.AcquireSingleton(config.ConfigUnmarshaler(nil))
	l := repository.NewLogrusLogger(
		cfg.GetString("scene.log.file"), cfg.GetInt("scene.log.max_size"),
		cfg.GetBool("scene.log.panic"),
	)
	l.SetLogLevel(logger.LogLevel(cfg.GetInt("scene.log.level")))
	registry.RegisterLogger(l.WithPrefix(cfg.GetString("scene.name")))
}

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

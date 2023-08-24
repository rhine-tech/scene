package builder

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/config"
	"github.com/aynakeya/scene/lens/infrastructure/ingestion/repository"
	"github.com/aynakeya/scene/registry"
)

// Init is instance of scene.LensInit
func Init() {
	cfg := registry.AcquireSingleton(config.ConfigUnmarshaler(nil))
	registry.Register(repository.NewOpenObserveCommonIngestor(
		cfg.GetString("openobserve.username"),
		cfg.GetString("openobserve.password"),
		cfg.GetString("openobserve.url"),
		cfg.GetString("openobserve.organization"),
	))
}

type DummyBuilder struct {
	scene.Builder
}

func (b DummyBuilder) Init() scene.LensInit {
	return func() {
		registry.Register(repository.NewDummyCommonIngestor())
	}
}

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

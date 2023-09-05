package builder

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask/repository"
	"github.com/rhine-tech/scene/registry"
)

// Init is instance of scene.LensInit
func Init() {
	registry.RegisterTaskDispatcher(repository.NewThunnusTaskDispatcher())
}

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

package builder

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/lens/infrastructure/asynctask/repository"
	"github.com/aynakeya/scene/registry"
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

package builder

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask/repository"
	"github.com/rhine-tech/scene/registry"
)

type Thunnus struct {
	scene.Builder
}

func (b Thunnus) Init() scene.LensInit {
	return func() {
		registry.RegisterTaskDispatcher(repository.NewThunnusTaskDispatcher())
	}
}

type Ants struct {
	scene.Builder
}

func (b Ants) Init() scene.LensInit {
	return func() {
		registry.RegisterTaskDispatcher(repository.NewAntsTaskDispatcher())
	}
}

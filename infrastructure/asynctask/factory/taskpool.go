package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask/taskpool"
	"github.com/rhine-tech/scene/registry"
)

type Thunnus struct {
	scene.ModuleFactory
}

func (b Thunnus) Init() scene.LensInit {
	return func() {
		registry.RegisterTaskDispatcher(taskpool.NewThunnusTaskDispatcher())
		registry.Register(taskpool.NewCommonCronTaskDispatcher(registry.TaskDispatcher))
	}
}

type Ants struct {
	scene.ModuleFactory
}

func (b Ants) Init() scene.LensInit {
	return func() {
		registry.RegisterTaskDispatcher(taskpool.NewAntsTaskDispatcher())
		registry.Register(taskpool.NewCommonCronTaskDispatcher(registry.TaskDispatcher))
	}
}

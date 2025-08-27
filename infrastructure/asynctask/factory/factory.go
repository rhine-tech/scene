package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/asynctask/repository"
	"github.com/rhine-tech/scene/infrastructure/asynctask/repository/longrunning"
	"github.com/rhine-tech/scene/registry"
)

type Thunnus struct {
	scene.ModuleFactory
}

func (b Thunnus) Init() scene.LensInit {
	return func() {
		registry.RegisterTaskDispatcher(repository.NewThunnusTaskDispatcher())
		registry.Register[asynctask.LongRunningTaskDispatcher](longrunning.NewLongRunningTaskDispatcher())
		registry.Register[asynctask.CronTaskDispatcher](repository.NewCommonCronTaskDispatcher(registry.TaskDispatcher))
	}
}

type Ants struct {
	scene.ModuleFactory
}

func (b Ants) Init() scene.LensInit {
	return func() {
		registry.RegisterTaskDispatcher(repository.NewAntsTaskDispatcher())
		registry.Register[asynctask.LongRunningTaskDispatcher](longrunning.NewLongRunningTaskDispatcher())
		registry.Register[asynctask.CronTaskDispatcher](repository.NewCommonCronTaskDispatcher(registry.TaskDispatcher))
	}
}

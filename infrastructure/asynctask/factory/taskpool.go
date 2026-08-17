package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/asynctask/taskpool"
	"github.com/rhine-tech/scene/registry"
)

type Thunnus struct {
	scene.ModuleFactory
}

func (b Thunnus) Init(container *registry.Container) {
	dispatcher := registry.Export[asynctask.TaskDispatcher](container, taskpool.NewThunnusTaskDispatcher())
	registry.Export[asynctask.CronTaskDispatcher](container, taskpool.NewCommonCronTaskDispatcher(dispatcher))
}

type Ants struct {
	scene.ModuleFactory
}

func (b Ants) Init(container *registry.Container) {
	dispatcher := registry.Export[asynctask.TaskDispatcher](container, taskpool.NewAntsTaskDispatcher())
	registry.Export[asynctask.CronTaskDispatcher](container, taskpool.NewCommonCronTaskDispatcher(dispatcher))
}

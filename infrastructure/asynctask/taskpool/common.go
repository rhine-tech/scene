package taskpool

import (
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/asynctask/taskpool/ants"
	"github.com/rhine-tech/scene/infrastructure/asynctask/taskpool/tuna"
)

func NewThunnusTaskDispatcher() asynctask.TaskDispatcher {
	return tuna.NewThunnus(asynctask.DefaultInitialPoolSize)
}

func NewAntsTaskDispatcher() asynctask.TaskDispatcher {
	return ants.NewAnts(asynctask.DefaultInitialPoolSize)
}

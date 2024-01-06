package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask"
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask/repository/ants"
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask/repository/tuna"
)

func NewThunnusTaskDispatcher() asynctask.TaskDispatcher {
	return tuna.NewThunnus(asynctask.DefaultInitialPoolSize)
}

func NewAntsTaskDispatcher() asynctask.TaskDispatcher {
	return ants.NewAnts(asynctask.DefaultInitialPoolSize)
}

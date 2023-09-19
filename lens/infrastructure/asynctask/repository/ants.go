package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask"
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask/repository/ants"
)

func NewAntsTaskDispatcher() asynctask.TaskDispatcher {
	return ants.NewAnts(asynctask.DefaultInitialPoolSize)
}

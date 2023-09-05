package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask"
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask/repository/tuna"
)

func NewThunnusTaskDispatcher() asynctask.TaskDispatcher {
	return tuna.NewThunnus(16)
}

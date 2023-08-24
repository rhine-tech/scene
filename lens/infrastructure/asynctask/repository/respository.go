package repository

import (
	"github.com/aynakeya/scene/lens/infrastructure/asynctask"
	"github.com/aynakeya/scene/lens/infrastructure/asynctask/repository/tuna"
)

func NewThunnusTaskDispatcher() asynctask.TaskDispatcher {
	return tuna.NewThunnus(16)
}

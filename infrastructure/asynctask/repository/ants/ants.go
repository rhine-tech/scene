package ants

import (
	xants "github.com/panjf2000/ants/v2"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
)

type AntsImpl struct {
	xants *xants.Pool
}

func (a AntsImpl) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("TaskDispatcher", "ants")
}

func (a AntsImpl) Run(task asynctask.TaskFunc) *asynctask.Task {
	t := &asynctask.Task{Func: task}
	a.RunTask(t)
	return t
}

func (a AntsImpl) RunTask(task *asynctask.Task) {
	task.SetStatus(asynctask.TaskStatusQueue)
	if err := a.xants.Submit(func() {
		task.SetStatus(asynctask.TaskStatusRunning)
		if err := task.Func(); err != nil {
			task.Err = err
		}
		task.SetStatus(asynctask.TaskStatusFinish)
	}); err != nil {
		task.Err = err
		task.SetStatus(asynctask.TaskStatusFinish)
	}
}

func NewAnts(poolSize int) *AntsImpl {
	p, err := xants.NewPool(poolSize)
	if err != nil {
		panic(err)
	}
	return &AntsImpl{
		xants: p,
	}
}

package longrunning

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"sync"
	"time"
)

type internalTask struct {
	ctx    context.Context
	cancel context.CancelFunc
	task   *asynctask.LongRunningTask
}

type defaultImpl struct {
	tasks sync.Map
}

func NewLongRunningTaskDispatcher() asynctask.LongRunningTaskDispatcher {
	return &defaultImpl{
		tasks: sync.Map{},
	}
}

func (d *defaultImpl) ImplName() scene.ImplName {
	return asynctask.Lens.ImplNameNoVer("LongRunningTaskDispatcher")
}

func (d *defaultImpl) Run(taskFunc asynctask.LongRunningTaskFunc) (task *asynctask.LongRunningTask, err error) {
	t := &asynctask.LongRunningTask{Func: taskFunc}
	taskId := t.Identifier()
	ctx, cancel := context.WithCancel(context.Background())
	d.tasks.Store(taskId, &internalTask{
		ctx:    ctx,
		cancel: cancel,
		task:   t,
	})
	go (func() {
		// remove task from entry
		defer d.tasks.Delete(taskId)
		for {
			select {
			case <-ctx.Done(): // Listen for context cancellation
				return
			default:
				func() {
					defer func() {
						if r := recover(); r != nil {
							//todo: handle panic
						}
					}()
					// todo handle error
					_ = t.Func(ctx)
				}()

				// If the context was cancelled during task execution, exit immediately
				if ctx.Err() != nil {
					return
				}
				// keep running man :)
				time.Sleep(1 * time.Second)
			}
		}
	})()
	return t, nil
}

func (d *defaultImpl) Cancel(id string) error {
	val, ok := d.tasks.Load(id)
	if !ok {
		return asynctask.ErrTaskNotFound
	}

	t, _ := val.(*internalTask)
	t.cancel()
	return nil
}

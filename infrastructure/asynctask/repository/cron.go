package repository

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/robfig/cron/v3"
	"sync"
)

type internalTask struct {
	task    *asynctask.CronTask
	entryId cron.EntryID
}

type CommonCronTaskDispatcher struct {
	cron           *cron.Cron
	taskDispatcher asynctask.TaskDispatcher
	tasks          sync.Map
	logger         logger.ILogger `aperture:""`
}

func (c *CommonCronTaskDispatcher) Dispose() error {
	c.cron.Stop()
	return nil
}

func (c *CommonCronTaskDispatcher) Setup() error {
	c.cron.Start()
	return nil
}

func (c *CommonCronTaskDispatcher) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("CronTaskDispatcher", "robfig_cron_v3")
}

func NewCommonCronTaskDispatcher(taskDispatcher asynctask.TaskDispatcher) asynctask.CronTaskDispatcher {
	return &CommonCronTaskDispatcher{
		cron:           cron.New(cron.WithSeconds()),
		taskDispatcher: taskDispatcher,
	}
}

func (c *CommonCronTaskDispatcher) Add(spec string, cmd asynctask.TaskFunc) (*asynctask.CronTask, error) {
	task := &asynctask.CronTask{
		Func: cmd,
	}
	return task, c.AddTask(spec, task)
}

func (c *CommonCronTaskDispatcher) AddWithName(spec string, name string, cmd asynctask.TaskFunc) (*asynctask.CronTask, error) {
	task := &asynctask.CronTask{
		Name: name,
		Func: cmd,
	}
	return task, c.AddTask(spec, task)
}

func (c *CommonCronTaskDispatcher) AddTask(spec string, task *asynctask.CronTask) error {
	entryId, err := c.cron.AddFunc(spec, func() {
		c.taskDispatcher.Run(func() error {
			c.logger.Infof("cron task %s start", task.Identifier())
			err := task.Func()
			if err != nil {
				task.ErrCount++
			}
			task.Total++
			if err != nil {
				c.logger.Errorf("cron task %s end with err = %v (run counts: %d)", task.Identifier(), err, task.Total)
			} else {
				c.logger.Infof("cron task %s end (run counts: %d)", task.Identifier(), task.Total)
			}
			return nil
		})
	})
	if err != nil {
		return err
	}
	c.tasks.Store(task.Identifier(), &internalTask{
		task:    task,
		entryId: entryId,
	})
	c.logger.Infof("cron task %s added with entryId = %d", task.Identifier(), entryId)
	return nil
}

func (c *CommonCronTaskDispatcher) Cancel(id string) error {
	val, ok := c.tasks.Load(id)
	if !ok {
		return asynctask.ErrTaskNotFound.WithDetailStr(id)
	}
	model, ok := val.(*internalTask)
	if !ok {
		c.logger.Errorf("cron task %s is not internal task, how's that possible", id)
		return asynctask.ErrInternal
	}
	c.cron.Remove(model.entryId)
	c.logger.Infof("cron task %s cancelled", id)
	return nil
}

func (c *CommonCronTaskDispatcher) GetTask(id string) (*asynctask.CronTask, error) {
	val, ok := c.tasks.Load(id)
	if !ok {
		return nil, asynctask.ErrTaskNotFound.WithDetailStr(id)
	}
	model, ok := val.(*internalTask)
	if !ok {
		c.logger.Errorf("cron task %s is not internal task, how's that possible", id)
		return nil, asynctask.ErrInternal
	}
	return model.task, nil
}

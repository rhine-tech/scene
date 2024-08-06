package repository

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/robfig/cron/v3"
	"time"
)

type CommonCronTaskDispatcher struct {
	cron           *cron.Cron
	taskDispatcher asynctask.TaskDispatcher
	tasks          map[string]*asynctask.CronTask
	logger         logger.ILogger `aperture:""`
}

func (c *CommonCronTaskDispatcher) Dispose() error {
	c.cron.Stop()
	return nil
}

func (c *CommonCronTaskDispatcher) Setup() error {
	c.logger = c.logger.WithPrefix(c.ImplName().String())
	c.cron.Start()
	return nil
}

func (c *CommonCronTaskDispatcher) ImplName() scene.ImplName {
	return asynctask.Lens.ImplName("CronTaskDispatcher", "cron_v3")
}

func NewCommonCronTaskDispatcher(taskDispatcher asynctask.TaskDispatcher) asynctask.CronTaskDispatcher {
	return &CommonCronTaskDispatcher{
		cron:           cron.New(cron.WithSeconds()),
		taskDispatcher: taskDispatcher,
		tasks:          make(map[string]*asynctask.CronTask),
	}
}

func (c *CommonCronTaskDispatcher) Add(spec string, cmd asynctask.TaskFunc) (*asynctask.CronTask, error) {
	task := &asynctask.CronTask{
		Func: cmd,
		Name: time.Now().String(),
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
	c.tasks[task.EntryID()] = task
	_, err := c.cron.AddFunc(spec, func() {
		c.taskDispatcher.Run(func() error {
			c.logger.Infof("cron task %s start", task.Name)
			err := task.Func()
			if err != nil {
				task.ErrCount++
			}
			task.Total++
			if err != nil {
				c.logger.Errorf("cron task %s end with err = %v (run counts: %d)", task.Name, err, task.Total)
			} else {
				c.logger.Infof("cron task %s end (run counts: %d)", task.Name, task.Total)
			}
			return nil
		})
	})
	return err
}

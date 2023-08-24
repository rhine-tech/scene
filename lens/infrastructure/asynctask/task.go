package asynctask

import (
	"sync/atomic"
	"time"
)

type TaskFunc func() error
type TaskStatus int32

const (
	TaskStatusQueue TaskStatus = iota
	TaskStatusRunning
	TaskStatusFinish
)

type Task struct {
	Func    TaskFunc
	Err     error
	status  int32
	timeout time.Duration
}

func (t *Task) SetStatus(status TaskStatus) {
	atomic.StoreInt32(&t.status, int32(status))
}

func (t *Task) Status() TaskStatus {
	return TaskStatus(atomic.LoadInt32(&t.status))
}

type TaskDispatcher interface {
	Run(task TaskFunc) *Task
	RunTask(task *Task)
}

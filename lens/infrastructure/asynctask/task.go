package asynctask

import (
	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
	"math"
	"sync/atomic"
	"time"
)

const DefaultMaxPoolSize = math.MaxInt32
const DefaultInitialPoolSize = 1024

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
	timeout time.Duration // not used
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

type CronTask struct {
	Name     string
	Func     TaskFunc
	Total    uint64
	ErrCount uint64
	id       string
}

func (t *CronTask) EntryID() string {
	if t.id != "" {
		return t.id
	}
	t.id = uuid.New().String()
	return t.id
}

type CronTaskDispatcher interface {
	scene.NamableImplementation
	Add(spec string, cmd TaskFunc) (*CronTask, error)
	AddWithName(spec string, name string, cmd TaskFunc) (*CronTask, error)
	AddTask(spec string, task *CronTask) error
}

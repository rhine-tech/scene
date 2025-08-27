package asynctask

import (
	"context"
	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
	"math"
	"sync/atomic"
	"time"
)

const Lens scene.InfraName = "asynctask"

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
	Name    string
	Func    TaskFunc
	Err     error
	status  int32
	timeout time.Duration // not used
}

func (t *Task) Identifier() string {
	if t.Name != "" {
		return t.Name
	}
	t.Name = uuid.New().String()
	return t.Name
}

func (t *Task) SetStatus(status TaskStatus) {
	atomic.StoreInt32(&t.status, int32(status))
}

func (t *Task) Status() TaskStatus {
	return TaskStatus(atomic.LoadInt32(&t.status))
}

type TaskDispatcher interface {
	scene.Named
	// Run run task once, the task should be ended in expected time
	Run(task TaskFunc) *Task
	// RunTask run task once, the task should be ended in expected time
	RunTask(task *Task)
}

type CronTask struct {
	Name        string // Use Identifier method to access
	Description string
	Func        TaskFunc
	Total       uint64
	ErrCount    uint64
}

// Identifier is the unique identifier getter, if this CronTask already
// set a name. the name will be the identifier
func (t *CronTask) Identifier() string {
	if t.Name != "" {
		return t.Name
	}
	t.Name = uuid.New().String()
	return t.Name
}

// CronTaskDispatcher is a service which will handle all cron task
type CronTaskDispatcher interface {
	scene.Named
	// Add will add task with a generated name common uuid
	Add(spec string, cmd TaskFunc) (*CronTask, error)
	// AddWithName will add task with specific name. name should be unique
	AddWithName(spec string, name string, cmd TaskFunc) (*CronTask, error)
	// AddTask is the underlying implementation for Add and AddWithName
	AddTask(spec string, task *CronTask) error
	// Cancel will cancel task have specific identifier
	Cancel(id string) error
	// GetTask will return the underlying task, not copy of the task info.
	// which means user can modify TaskFunc if they need
	GetTask(id string) (*CronTask, error)
}

// LongRunningTaskFunc is a function which handle long-running task,
// this function must return when context is done
type LongRunningTaskFunc func(ctx context.Context) error

type LongRunningTask struct {
	Name        string
	Description string
	Func        LongRunningTaskFunc
}

// Identifier is the unique identifier getter, if this LongRunningTask already
// set a name. the name will be the identifier
func (t *LongRunningTask) Identifier() string {
	if t.Name != "" {
		return t.Name
	}
	t.Name = uuid.New().String()
	return t.Name
}

type LongRunningTaskDispatcher interface {
	scene.Named
	Run(taskFunc LongRunningTaskFunc) (task *LongRunningTask, err error)
	//RunWithName(name string, taskFunc LongRunningTaskFunc) (task *LongRunningTask, err error)
	Cancel(id string) error
}

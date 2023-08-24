package tuna

import (
	"github.com/aynakeya/scene/lens/infrastructure/asynctask"
	"github.com/aynakeya/scene/pkg/queue"
	"runtime"
)

type Thunnus struct {
	taskQueue *queue.QueueChannel[*asynctask.Task]
	taskChan  chan *asynctask.Task
	tunas     []*tuna
}

func NewThunnus(concurs int) *Thunnus {
	ch := make(chan *asynctask.Task, 128)
	tns := &Thunnus{
		taskQueue: queue.NewQueueChannelFromChan(ch),
		taskChan:  ch,
		tunas:     make([]*tuna, concurs),
	}
	for i := 0; i < concurs; i++ {
		tns.tunas[i] = newTuna(ch)
		go tns.tunas[i].run()
	}
	return tns
}

func (t *Thunnus) Run(tsk asynctask.TaskFunc) *asynctask.Task {
	task := &asynctask.Task{Func: tsk}
	t.RunTask(task)
	return task
}

func (t *Thunnus) RunTask(task *asynctask.Task) {
	t.taskQueue.Push(task)
}

func (t *Thunnus) Stop() {
	t.Resize(0)
}

// Resize reference to https://github.com/Jeffail/tunny/blob/master/tunny.go#L266
func (t *Thunnus) Resize(n int) {
	cnt := len(t.tunas)
	if cnt == n {
		return
	}

	// Add extra workers if N > len(workers)
	for i := cnt; i < n; i++ {
		t.tunas = append(t.tunas, newTuna(t.taskChan))
	}

	// Asynchronously stop all workers > N
	for i := n; i < cnt; i++ {
		t.tunas[i].stop()
	}

	// Synchronously wait for all workers > N to stop
	for i := n; i < cnt; i++ {
		t.tunas[i].join()
		t.tunas[i] = nil
	}
	t.tunas = t.tunas[:n]
}

type tuna struct {
	taskChan chan *asynctask.Task
	stopSig  chan int8
	finished chan int8
}

func newTuna(taskChan chan *asynctask.Task) *tuna {
	return &tuna{
		taskChan: taskChan,
		stopSig:  make(chan int8),
		finished: make(chan int8),
	}
}

func (w *tuna) run() {
	for {
		select {
		case <-w.stopSig:
			return
		case tsk, ok := <-w.taskChan:
			if !ok {
				goto finish
			}
			tsk.Err = tsk.Func()
			tsk.SetStatus(asynctask.TaskStatusFinish)
		default:
			if w.taskChan == nil {
				goto finish
			}
			runtime.Gosched()
		}
	}
finish:
	w.finished <- 1
}

func (w *tuna) stop() {
	close(w.stopSig)
}

func (w *tuna) join() {
	<-w.finished
}

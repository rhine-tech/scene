package repository

import (
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"sync"
	"testing"
)

func testTaskDispatcher(tp asynctask.TaskDispatcher) {
	var wg sync.WaitGroup
	for i := 0; i < 20000000; i++ {
		wg.Add(1)
		tp.Run(func() error {
			wg.Done()
			return nil
		})
	}
	wg.Wait()
}

func TestTaskDispatcherCommon(t *testing.T) {
	testTaskDispatcher(NewThunnusTaskDispatcher())
	t.Log("ThunnusTaskDispatcher test passed")
	testTaskDispatcher(NewAntsTaskDispatcher())
	t.Log("AntsTaskDispatcher test passed")
}

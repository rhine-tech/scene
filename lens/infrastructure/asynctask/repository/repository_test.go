package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask"
	"sync"
	"testing"
)

func testTaskDispatcher(tp asynctask.TaskDispatcher) {
	var wg sync.WaitGroup
	for i := 0; i < 32767; i++ {
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
	testTaskDispatcher(NewAntsTaskDispatcher())
}

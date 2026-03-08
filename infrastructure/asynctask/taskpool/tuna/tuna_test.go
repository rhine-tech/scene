package tuna

import (
	"sync"
	"testing"
)

func TestThunnus_Run(t *testing.T) {
	thunnus := NewThunnus(100)
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		thunnus.Run(func() error {
			wg.Done()
			return nil
		})
	}
	wg.Wait()
}

func TestThunnus_Resize(t *testing.T) {
	thunnus := NewThunnus(10)
	thunnus.Resize(100)
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		thunnus.Run(func() error {
			wg.Done()
			return nil
		})
	}
	wg.Wait()
}

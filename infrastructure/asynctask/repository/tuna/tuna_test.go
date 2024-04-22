package tuna

import (
	"fmt"
	"sync"
	"testing"
)

func TestThunnus_Run(t *testing.T) {
	thunnus := NewThunnus(16)
	var wg sync.WaitGroup
	for i := 0; i < 20000000; i++ {
		wg.Add(1)
		thunnus.Run(func() error {
			//v := time.Second * time.Duration(rand.Intn(3))
			//fmt.Println("sleep", v)
			//time.Sleep(v)
			wg.Done()
			return nil
		})
	}
	fmt.Println("wait")
	wg.Wait()
	thunnus.Stop()
	fmt.Println("done")
}

func TestThunnus_RunSeq(t *testing.T) {
	thunnus := NewThunnus(4)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		v := i
		thunnus.Run(func() error {
			fmt.Println(v)
			wg.Done()
			return nil
		})
	}
	fmt.Println("wait")
	wg.Wait()
	thunnus.Stop()
	fmt.Println("done")
}

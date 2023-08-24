package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestQueueChannel(t *testing.T) {
	qc := NewQueueChannel[int](128)
	var wg sync.WaitGroup
	wg.Add(1024 * 1234)
	go func() {
		for x := range qc.ch {
			if x == -1 {
				t.Errorf("Fail")
			}
			wg.Done()
		}
	}()

	for i := 0; i < 1024; i++ {
		go func() {
			for j := 0; j < 1234; j++ {
				qc.Push(j)
			}
		}()
	}
	fmt.Println("Finish add")
	wg.Wait()
}

func TestQueueChannel1(t *testing.T) {
	qc := NewQueueChannel[int](128)
	var wg sync.WaitGroup
	wg.Add(1024 * 2)
	for i := 0; i < 10; i++ {
		ii := i
		go func() {
			for x := range qc.ch {
				fmt.Printf("%d done in worker %d\n", x, ii)
				wg.Done()
				//time.Sleep(time.Second * time.Duration(rand.Intn(3)+2))
			}
		}()
	}

	time.Sleep(time.Second * 10)
	fmt.Println("start adding")
	for j := 0; j < 1024; j++ {
		qc.Push(j)
	}
	fmt.Println("adding finished")
	time.Sleep(time.Second * 10)
	fmt.Println("start adding")
	for j := 0; j < 1024; j++ {
		qc.Push(j)
	}
	fmt.Println("adding finished")
	wg.Wait()
}

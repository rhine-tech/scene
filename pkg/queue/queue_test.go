package queue

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestQueue(t *testing.T) {
	queue := NewQueue[int]()
	for i := 0; i < 150; i++ {
		queue.Push(i)
		assert.Equal(t, queue.Count(), i+1)
		if i == 15 {
			assert.Equal(t, queue.Size(), 32)
		}
		if i == 31 {
			assert.Equal(t, queue.Size(), 64)
		}
	}
	for i := 0; i < 150; i++ {
		x := queue.Pop()
		//fmt.Println(x)
		assert.Equal(t, x, i)
	}

	assert.Equal(t, queue.Size(), 16)
}

func TestQueue2(t *testing.T) {
	queue := NewQueue[int]()
	for i := 0; i < 10; i++ {
		queue.Push(i)
	}
	for i := 0; i < 10; i++ {
		queue.Pop()
	}
	for i := 0; i < 10; i++ {
		queue.Push(i)
	}
	for i := 0; i < 10; i++ {
		queue.Pop()
	}

	for i := 0; i < 150; i++ {
		queue.Push(i)
		assert.Equal(t, queue.Count(), i+1)
		if i == 15 {
			assert.Equal(t, queue.Size(), 32)
		}
		if i == 31 {
			assert.Equal(t, queue.Size(), 64)
		}
	}
	for i := 0; i < 150; i++ {
		x := queue.Pop()
		//fmt.Println(x)
		assert.Equal(t, x, i)
	}

	assert.Equal(t, queue.Size(), 16)
}

func TestQueueConcurrencySafe(t *testing.T) {
	queue := NewQueue[int]()
	var wg sync.WaitGroup
	for i := 0; i < 1024; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1234; j++ {
				queue.Push(j)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, queue.Count(), 1024*1234)
}

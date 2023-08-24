package queue

import (
	"runtime"
)

type QueueChannel[T any] struct {
	ch    chan T
	input chan int
	size  int
	cache *Queue[T]
}

func NewQueueChannel[T any](size int) *QueueChannel[T] {
	return NewQueueChannelFromChan(make(chan T, size))
}

func NewQueueChannelFromChan[T any](ch chan T) *QueueChannel[T] {
	qc := &QueueChannel[T]{
		ch:    ch,
		input: make(chan int, 1),
		size:  cap(ch),
		cache: NewQueue[T](),
	}
	go qc.magic()
	return qc
}

func (q *QueueChannel[T]) Stop() {
	if q.ch == nil {
		return
	}
	close(q.ch)
	q.ch = nil
}

func (q *QueueChannel[T]) Size() int {
	return q.size
}

func (q *QueueChannel[T]) Chan() chan T {
	return q.ch
}

func (q *QueueChannel[T]) Push(elem T) {
	select {
	case q.ch <- elem:
	default:
		q.cache.Push(elem)
	}
}

func (q *QueueChannel[T]) magic() {
	for q.ch != nil {
		if q.cache.Count() > 0 {
			q.ch <- q.cache.Pop()
		} else {
			runtime.Gosched()
		}
	}
}

package queue

import "sync"

const minQueueSize = 16

// Queue A simple thread safe resizable queue.
type Queue[T any] struct {
	queue []T
	front int
	back  int
	count int
	// make it thread safe, not considering time issue for now
	lock sync.RWMutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		queue: make([]T, minQueueSize),
	}
}

// Size return current size of the queue
func (q *Queue[T]) Size() int {
	return len(q.queue)
}

// Count return the current element count in the queue
func (q *Queue[T]) Count() int {
	return q.count
}

// Front return the element in th front of the queue
func (q *Queue[T]) Front() T {
	return q.queue[q.front]
}

// Back return the element in the back of the queue
func (q *Queue[T]) Back() T {
	return q.queue[q.back]
}

// Pop pop the front element and return
func (q *Queue[T]) Pop() T {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.count == 0 {
		return *new(T)
	}
	elem := q.queue[q.front]
	q.queue[q.front] = *new(T)
	q.count--
	// work for all size
	//q.front = (q.front + 1) % size
	// works for size = 2^n
	q.front = (q.front + 1) & (q.Size() - 1)
	if q.Size() > minQueueSize && q.count < q.Size()/4 {
		q.resize(q.Size() / 2)
	}
	return elem
}

// Elems return a copy of the queue.
func (q *Queue[T]) Elems() []T {
	q.lock.Lock()
	defer q.lock.Unlock()
	elems := make([]T, q.count)
	if q.front < q.back {
		copy(elems, q.queue[q.front:q.back])
	} else {
		// copy end of queue to the front
		n := copy(elems, q.queue[q.front:])
		// copy rest of element
		copy(elems[n:], q.queue[:q.back])
	}
	return elems
}

// Push push the element at the back the queue
func (q *Queue[T]) Push(elem T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.count+1 == q.Size() {
		q.resize(q.Size() * 2)
	}
	q.queue[q.back] = elem
	q.back = (q.back + 1) & (q.Size() - 1)
	q.count++
}

// resize is internal method for resizing
func (q *Queue[T]) resize(size int) {
	newq := make([]T, size)
	if q.front < q.back {
		copy(newq, q.queue[q.front:q.back])
	} else {
		// copy end of queue to the front
		n := copy(newq, q.queue[q.front:])
		// copy rest of element
		copy(newq[n:], q.queue[:q.back])
	}

	// make front to 0
	q.front = 0
	// the back for the queue is now the element count
	q.back = q.count
	q.queue = newq
}

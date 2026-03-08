package queue

import (
	"github.com/rhine-tech/scene/infrastructure/asynctask/queue/memoryqueue"
)

func NewMemoryTaskQueue() *memoryqueue.MemoryQueue {
	return memoryqueue.NewMemoryQueue()
}

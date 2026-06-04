package memoryreporter

import "github.com/rhine-tech/scene/infrastructure/asynctask/reporter/memoryreporter/v2"

type Config = v2.Config
type MemoryReporter = v2.MemoryReporter

func NewMemoryReporter(config ...Config) *MemoryReporter {
	return v2.NewMemoryReporter(config...)
}

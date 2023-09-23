package registry

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"reflect"
	"sort"
	"sync"
)

type Registry[I comparable, T any] interface {
	Register(impl T)
	RegisterByEntry(entry I, impl T)
	Acquire(impl T) T
	AcquireByEntry(entry I) T
	AcquireAll() []T
}

type naming[I comparable, T any] func(value T) I

func indexedNaming[T any]() naming[int, T] {
	index := -1
	return func(value T) int {
		index++
		return index
	}
}

type registry[I comparable, T any] struct {
	registry map[I]T
	lock     sync.RWMutex
	naming   naming[I, T]
}

func NewRegistry[I comparable, T any](naming naming[I, T]) Registry[I, T] {
	return &registry[I, T]{
		registry: make(map[I]T),
		naming:   naming,
	}
}

func (r *registry[I, T]) Register(impl T) {
	if r.naming == nil {
		panic("naming function not set")
	}
	r.RegisterByEntry(r.naming(impl), impl)
}

func (r *registry[I, T]) RegisterByEntry(entry I, impl T) {
	r.lock.Lock()
	if _, ok := r.registry[entry]; ok {
		panic(fmt.Sprintf("duplicate registry entry %v in %s", entry, getInterfaceName[T]()))
	}
	r.registry[entry] = impl
	r.lock.Unlock()
}

func (r *registry[I, T]) Acquire(impl T) T {
	return r.AcquireByEntry(r.naming(impl))
}

func (r *registry[I, T]) AcquireByEntry(entry I) T {
	r.lock.RLock()
	val, ok := r.registry[entry]
	r.lock.RUnlock()
	if !ok {
		panic(fmt.Sprintf("no registry entry %v", entry))
	}
	return val
}

func (r *registry[I, T]) AcquireAll() []T {
	r.lock.RLock()
	sets := make(map[uintptr]uint8, 0)
	vals := make([]T, 0, len(r.registry))
	for _, val := range r.registry {
		if _, ok := sets[reflect.ValueOf(val).Pointer()]; ok {
			continue
		}
		vals = append(vals, val)
		sets[reflect.ValueOf(val).Pointer()] = 1
	}
	r.lock.RUnlock()
	return vals
}

type orderedregistry[I constraints.Ordered, T any] struct {
	*registry[I, T]
}

func (r *orderedregistry[I, T]) Register(impl T) {
	if r.naming == nil {
		panic("naming function not set")
	}
	r.RegisterByEntry(r.naming(impl), impl)
}

func (r *orderedregistry[I, T]) RegisterByEntry(entry I, impl T) {
	r.lock.Lock()
	if _, ok := r.registry.registry[entry]; ok {
		panic(fmt.Sprintf("duplicate registry entry %v in %s", entry, getInterfaceName[T]()))
	}

	r.registry.registry[entry] = impl
	r.lock.Unlock()
}

func (r *orderedregistry[I, T]) AcquireAll() []T {
	r.lock.RLock()
	sets := make(map[uintptr]uint8, 0)
	keys := make([]I, 0, len(r.registry.registry))
	for key := range r.registry.registry {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	vals := make([]T, 0, len(r.registry.registry))
	for _, key := range keys {
		if _, ok := sets[reflect.ValueOf(r.registry.registry[key]).Pointer()]; ok {
			continue
		}
		vals = append(vals, r.registry.registry[key])
		sets[reflect.ValueOf(r.registry.registry[key]).Pointer()] = 1
	}
	r.lock.RUnlock()
	return vals
}

func NewOrderedRegistry[I constraints.Ordered, T any](naming naming[I, T]) Registry[I, T] {
	return &orderedregistry[I, T]{
		registry: &registry[I, T]{
			registry: make(map[I]T),
			naming:   naming,
		},
	}
}

type Registrant func(impl any)

func registrantWrapper[I comparable, T any](reg Registry[I, T]) Registrant {
	return func(impl any) {
		implT, ok := impl.(T)
		if !ok {
			return
		}
		reg.Register(implT)
	}
}

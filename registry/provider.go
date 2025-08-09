package registry

import "sync"

// todo: finish me

type Dependency[T any] interface {
	Inject(T)
}

type Provider[T any] interface {
	Provide() T
}

type SingletonProvider[T any] struct {
	once     sync.Once
	instance T
}

func NewSingletonProvider[T any](instance T) *SingletonProvider[T] {
	return &SingletonProvider[T]{
		instance: instance,
	}
}

func (p *SingletonProvider[T]) Provide() T {
	p.once.Do(func() {
		p.instance = Use(p.instance)
	})
	return p.instance
}

type Container struct {
	name         string
	dependencies map[string]any
}

func (c *Container) ContainerName() string {
	return c.name
}

func NewContainer(providers ...Provider[any]) *Container {
	return &Container{
		dependencies: make(map[string]any),
	}
}

func ContainerProvide[T any](c *Container) (T, bool) {
	v, ok := c.dependencies[getInterfaceName[T]()]
	if !ok {
		return *new(T), false
	}
	return v.(T), true
}

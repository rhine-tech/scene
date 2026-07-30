package registry

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/rhine-tech/scene"
)

type Container struct {
	lock              sync.RWMutex
	dependencies      map[string]any
	hooks             map[string][]InjectHookFunc
	pendingInjections []any
	disposable        Registry[int, scene.Disposable]
	setupable         Registry[int, scene.Setupable]
	registrants       []Registrant
}

var defaultContainer = NewContainer()

func NewContainer() *Container {
	disposable := NewOrderedRegistry(indexedNaming[scene.Disposable]())
	setupable := NewOrderedRegistry(indexedNaming[scene.Setupable]())
	return &Container{
		dependencies: make(map[string]any),
		hooks:        make(map[string][]InjectHookFunc),
		disposable:   disposable,
		setupable:    setupable,
		registrants: []Registrant{
			registrantWrapper(disposable),
			registrantWrapper(setupable),
		},
	}
}

func ContainerProvide[T any](c *Container, name ...string) T {
	key := dependencyName[T](name)
	v, ok := c.lookup(key)
	if !ok {
		panic(fmt.Sprintf("no dependency registered for %s", key))
	}
	return v.(T)
}

// ContainerRegister registers and injects a value in c.
func ContainerRegister[T any](c *Container, val T, name ...string) T {
	key := dependencyName[T](name)
	for _, registrant := range c.registrants {
		registrant(val)
	}
	c.lock.Lock()
	c.dependencies[key] = val
	c.lock.Unlock()
	if !c.queuePendingInjection(val) {
		ContainerInject(c, val)
	}
	return val
}

// ContainerMustRegister registers val in c or panics when err is non-nil.
func ContainerMustRegister[T any](c *Container, val T, err error, name ...string) T {
	if err != nil {
		panic(err)
	}
	return ContainerRegister(c, val, name...)
}

// ContainerUse returns val when it is usable, otherwise it resolves T from c.
func ContainerUse[T any](c *Container, val T) T {
	if canUse(val) {
		return val
	}
	return ContainerProvide[T](c)
}

// ContainerLoad injects val and registers its lifecycle interfaces without
// registering val as a dependency.
func ContainerLoad[T any](c *Container, val T) T {
	for _, registrant := range c.registrants {
		registrant(val)
	}
	if !c.queuePendingInjection(val) {
		return ContainerInject(c, val)
	}
	return val
}

// ContainerInject injects dependencies from c into injectable.
func ContainerInject[T any](c *Container, injectable T) T {
	val := reflect.ValueOf(injectable)
	indirectVal := reflect.Indirect(val) // In case injectable is a pointer
	inject[T](c, indirectVal)
	return injectable
}

func dependencyName[T any](name []string) string {
	switch len(name) {
	case 0:
		return getInterfaceName[T]()
	case 1:
		if name[0] == "" {
			return getInterfaceName[T]()
		}
		return name[0]
	default:
		panic("scene registry: dependency accepts at most one name")
	}
}

func (c *Container) lookup(name string) (interface{}, bool) {
	c.lock.RLock()
	impl, ok := c.dependencies[name]
	c.lock.RUnlock()
	return impl, ok
}

func (c *Container) queuePendingInjection(val any) bool {
	c.lock.Lock()
	if c.pendingInjections == nil {
		c.lock.Unlock()
		return false
	}
	c.pendingInjections = append(c.pendingInjections, val)
	c.lock.Unlock()
	return true
}

func (c *Container) beginLazyInjection() {
	c.lock.Lock()
	if c.pendingInjections != nil {
		c.lock.Unlock()
		panic("scene registry: lazy injection is already active for this container")
	}
	c.pendingInjections = make([]any, 0)
	c.lock.Unlock()
}

func (c *Container) finishLazyInjection() []any {
	c.lock.Lock()
	pending := c.pendingInjections
	c.pendingInjections = nil
	c.lock.Unlock()
	return pending
}

func (c *Container) abortLazyInjection() {
	c.lock.Lock()
	c.pendingInjections = nil
	c.lock.Unlock()
}

// WithLazyInjection delays injection in c until proc returns.
// It must not be nested, and proc must perform registration synchronously.
func (c *Container) WithLazyInjection(proc func()) {
	c.beginLazyInjection()
	finished := false
	defer func() {
		if !finished {
			c.abortLazyInjection()
		}
	}()
	proc()
	pending := c.finishLazyInjection()
	finished = true
	for _, val := range pending {
		ContainerInject(c, val)
	}
}

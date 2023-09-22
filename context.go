package scene

import (
	"sync"
)

type Context interface {
	Get(key string) (value any, exists bool)
	Set(key string, value any)
}

type defaultCtx struct {
	values map[string]any
	mux    sync.RWMutex
}

func NewContext() Context {
	return &defaultCtx{}
}

func (c *defaultCtx) Set(key string, value any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.values == nil {
		c.values = make(map[string]any)
	}
	c.values[key] = value
}

func (c *defaultCtx) Get(key string) (value any, exists bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	value, exists = c.values[key]
	return value, exists
}

func ContextSetValue[T any](ctx Context, value T) {
	ctx.Set(GetInterfaceName[T](), value)
}

func ContextFindValue[T any](ctx Context) (T, bool) {
	v, ok := ctx.Get(GetInterfaceName[T]())
	if !ok {
		return *new(T), ok
	}
	valueT, ok := v.(T)
	return valueT, ok
}

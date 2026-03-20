package scene

import (
	"context"
	"sync"
	"time"
)

type MutableContext interface {
	SetContextValue(key any, value any)
}

type mutableContext struct {
	base   context.Context
	values map[any]any
	mux    sync.RWMutex
}

func NewContext() context.Context {
	return &mutableContext{
		base:   context.Background(),
		values: make(map[any]any),
	}
}

func (c *mutableContext) Deadline() (deadline time.Time, ok bool) {
	return c.base.Deadline()
}

func (c *mutableContext) Done() <-chan struct{} {
	return c.base.Done()
}

func (c *mutableContext) Err() error {
	return c.base.Err()
}

func (c *mutableContext) Value(key any) any {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if value, exists := c.values[key]; exists {
		return value
	}
	return c.base.Value(key)
}

func (c *mutableContext) SetContextValue(key any, value any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.values[key] = value
}

func ContextSetValue[T any](ctx context.Context, key any, value T) context.Context {
	if setter, ok := ctx.(MutableContext); ok {
		setter.SetContextValue(key, value)
		return ctx
	}
	return context.WithValue(ctx, key, value)
}

func ContextFindValue[T any](ctx context.Context, key any) (T, bool) {
	v := ctx.Value(key)
	if v == nil {
		return *new(T), false
	}
	valueT, ok := v.(T)
	return valueT, ok
}

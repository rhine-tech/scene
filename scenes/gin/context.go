package gin

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Context[T any] struct {
	*gin.Context
	App T // App is the container of current app
}

func (g *Context[T]) Deadline() (deadline time.Time, ok bool) {
	if g.Request == nil {
		return time.Time{}, false
	}
	return g.Request.Context().Deadline()
}

func (g *Context[T]) Done() <-chan struct{} {
	if g.Request == nil {
		return nil
	}
	return g.Request.Context().Done()
}

func (g *Context[T]) Err() error {
	if g.Request == nil {
		return nil
	}
	return g.Request.Context().Err()
}

func (g *Context[T]) Value(key any) any {
	if val, exists := g.Context.Get(contextStorageKey(key)); exists {
		return val
	}
	if g.Request != nil {
		if val := g.Request.Context().Value(key); val != nil {
			return val
		}
	}
	return nil
}

func (g *Context[T]) SetContextValue(key any, value any) {
	g.Context.Set(contextStorageKey(key), value)
}

func contextStorageKey(key any) string {
	return "scene.ctx." + fmt.Sprintf("%T:%v", key, key)
}

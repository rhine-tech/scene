package arpc

import (
	"context"
	"fmt"
	"time"

	"github.com/lesismal/arpc"
)

type arpcContext struct {
	val *arpc.Context
}

func (a *arpcContext) Deadline() (deadline time.Time, ok bool) {
	return a.val.Deadline()
}

func (a *arpcContext) Done() <-chan struct{} {
	return a.val.Done()
}

func (a *arpcContext) Err() error {
	return a.val.Err()
}

func (a *arpcContext) Value(key any) any {
	value, _ := a.val.Get(contextStorageKey(key))
	return value
}

func (a *arpcContext) SetContextValue(key any, value any) {
	a.val.Set(contextStorageKey(key), value)
}

func contextStorageKey(key any) string {
	return "scene.ctx." + fmt.Sprintf("%T:%v", key, key)
}

// Context this context is not fully supported yet
func Context(ctx *arpc.Context) context.Context {
	return &arpcContext{ctx}
}

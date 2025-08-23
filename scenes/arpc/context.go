package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene"
)

type arpcContext struct {
	val *arpc.Context
}

func (a *arpcContext) Get(key string) (value any, exists bool) {
	return a.val.Get(key)
}

func (a *arpcContext) Set(key string, value any) {
	a.val.Set(key, value)
}

func Context(ctx *arpc.Context) scene.Context {
	return &arpcContext{ctx}
}

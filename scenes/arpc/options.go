package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/util"
	"github.com/rhine-tech/scene/registry"
)

func UseRecover() ServerOption {
	return func(_ *registry.Scope, server *arpc.Server) error {
		server.Handler.Use(func(ctx *arpc.Context) {
			defer util.Recover()
			ctx.Next()
		})
		return nil
	}
}

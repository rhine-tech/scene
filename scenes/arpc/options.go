package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/util"
)

func UseRecover() ARpcOption {
	return func(server *arpc.Server) error {
		server.Handler.Use(func(ctx *arpc.Context) {
			defer util.Recover()
			ctx.Next()
		})
		return nil
	}
}

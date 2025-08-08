package arpc

import (
	"errors"
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/utils/must"
	"time"
)

const rpcNameBasicPasswordAuth = "scene.arpc.auth.password"

func WithPassword(password string) ClientOption {
	return func(client *arpc.Client) error {
		log := must.Must(client.Get("logger.ILogger")).(logger.ILogger)
		client.Handler.HandleConnected(func(c *arpc.Client) {
			var data string
			err := c.Call(rpcNameBasicPasswordAuth, &password, &data, time.Second*5)
			if err != nil {
				log.Error("WithPassword: fail to call scene.arpc.auth.password", err.Error())
				return
			}
			if data == "ok" {
				log.Info("WithPassword: auth with password success")
				return
			}
			log.Warn("WithPassword: auth with password failed, reason:", data)
		})
		return nil
	}
}

func UsePassword(password string) ARpcOption {
	return func(server *arpc.Server) error {
		server.Handler.Use(func(ctx *arpc.Context) {
			method := ctx.Message.Method()
			if method == rpcNameBasicPasswordAuth {
				ctx.Next()
				return
			}

			_, authenticated := ctx.Client.Get("scene.arpc.password")
			if !authenticated {
				ctx.Error(errors.New("connection not authenticated"))
				ctx.Abort()
				return
			}

			ctx.Next()
		})
		server.Handler.Handle(rpcNameBasicPasswordAuth, func(context *arpc.Context) {
			var data string
			err := context.Bind(&data)
			if err != nil {
				_ = context.Write("fail")
				return
			}
			if data == password {
				context.Client.Set("scene.arpc.password", password)
				_ = context.Write("ok")
			} else {
				_ = context.Write("fail")
			}
		})
		return nil
	}
}

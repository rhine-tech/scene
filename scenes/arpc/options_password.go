package arpc

import (
	"errors"
	"github.com/lesismal/arpc"
	"time"
)

const rpcNameBasicPasswordAuth = "scene.arpc.auth.password"

func WithPassword(password string) ClientOption {
	return func(client Client) error {
		client.AddConnectedHandler(func(c *arpc.Client) {
			var data string
			err := c.Call(rpcNameBasicPasswordAuth, &password, &data, time.Second*5)
			if err != nil {
				client.Logger().Error("WithPassword: fail to call scene.arpc.auth.password", err.Error())
				return
			}
			if data == "ok" {
				client.Logger().Info("WithPassword: auth with password success")
				return
			}
			client.Logger().Warn("WithPassword: auth with password failed, reason:", data)
		})
		return nil
	}
}

func UsePassword(password string) ServerOption {
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

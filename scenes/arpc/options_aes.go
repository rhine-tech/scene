package arpc

import (
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/scenes/arpc/helper"

	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/utils/must"
)

// --- Client Side ---

// WithAesEncryption enables automatic AES encryption for all messages on the client after a successful handshake.
func WithAesEncryption(key []byte) ClientOption {
	return func(client *arpc.Client) error {
		log := must.Must(client.Get("logger.ILogger")).(logger.ILogger)
		coder := helper.NewAesCoder(key, log)
		client.Handler.UseCoder(coder)
		return nil
	}
}

// --- Server Side ---

// UseAesEncryption enables automatic AES decryption for all messages after a successful handshake.
func UseAesEncryption(key []byte) ARpcOption {
	return func(server *arpc.Server) error {
		log := registry.Logger.WithPrefix((&arpcContainer{}).ImplName().Identifier())
		coder := helper.NewAesCoder(key, log)
		server.Handler.UseCoder(coder)
		return nil
	}
}

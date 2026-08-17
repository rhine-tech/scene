package arpc

import (
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/scenes/arpc/helper"

	"github.com/lesismal/arpc"
)

// --- Client Side ---

// WithAesEncryption enables automatic AES encryption for all messages on the client after a successful handshake.
func WithAesEncryption(key []byte) ClientOption {
	return func(client Client) error {
		coder := helper.NewAesCoder(key, client.Logger())
		client.Client().Handler.UseCoder(coder)
		return nil
	}
}

// --- Server Side ---

// UseAesEncryption enables automatic AES decryption for all messages after a successful handshake.
func UseAesEncryption(key []byte) ServerOption {
	return func(scope *registry.Scope, server *arpc.Server) error {
		log, exists := registry.LookupIn[logger.ILogger](scope)
		if !exists {
			log = logger.NoopLogger{}
		}
		log = log.WithPrefix((&arpcContainer{}).ImplName().Identifier())
		coder := helper.NewAesCoder(key, log)
		server.Handler.UseCoder(coder)
		return nil
	}
}

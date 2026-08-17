package arpc

import (
	"github.com/lesismal/arpc"
	"github.com/rhine-tech/scene/registry"
)

type ServerOption func(scope *registry.Scope, server *arpc.Server) error

// ClientOption is initialization option for arpc Client
type ClientOption func(server Client) error

package arpc

import (
	"github.com/lesismal/arpc"
)

type ARpcOption func(server *arpc.Server) error

// ClientOption is initialization option for arpc Client
type ClientOption func(server *arpc.Client) error

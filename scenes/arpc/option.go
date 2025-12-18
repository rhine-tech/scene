package arpc

import (
	"github.com/lesismal/arpc"
)

type ServerOption func(server *arpc.Server) error

// ClientOption is initialization option for arpc Client
type ClientOption func(server Client) error

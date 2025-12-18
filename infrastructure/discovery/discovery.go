package discovery

import (
	"context"
	"fmt"

	"github.com/rhine-tech/scene"
)

type Endpoint struct {
	Addr string
}

type Node struct {
	Key    string
	Server string
}

func (n *Node) String() string {
	return fmt.Sprintf("<DiscoveryNode %s>", n.Key)
}

type Registerer interface {
	scene.Named
	Register(name string, endpoint Endpoint) (*Node, error)
	Deregister(node *Node) error
}

// Resolver returns available endpoints for a given service key.
type Resolver interface {
	Resolve(ctx context.Context, key string) ([]Endpoint, error)
	// Watch registers a callback which will be invoked whenever the endpoints change.
	// The returned cancel func should be called to stop watching.
	Watch(ctx context.Context, key string, handler EndpointWatchHandler) (context.CancelFunc, error)
}

// EndpointWatchHandler consumes the latest endpoint snapshot for a service key.
type EndpointWatchHandler func([]Endpoint)

package discovery

import (
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

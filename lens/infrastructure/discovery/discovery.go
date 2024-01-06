package discovery

import "fmt"

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

type Register interface {
	Register(name string, endpoint Endpoint) (*Node, error)
	Deregister(node *Node) error
}

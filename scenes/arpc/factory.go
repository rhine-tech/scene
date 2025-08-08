package arpc

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/registry"
)

type ClientProvider scene.IModuleDependencyProvider[Client]

type ClientFactory struct {
	scene.ModuleFactory
	Network string
	Addr    string
	Options []ClientOption
}

func (c ClientFactory) Init() scene.LensInit {
	return func() {
		registry.Register[Client](NewClient(c.Network, c.Addr, c.Options...))
	}
}

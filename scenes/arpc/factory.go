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

type NexusClientFactory struct {
	scene.ModuleFactory
	Network string
	Addr    string
	Options []ClientOption
	UseApps []scene.AppInit[ARpcApp]
}

func (c NexusClientFactory) Init() scene.LensInit {
	return func() {
		apps := make([]ARpcApp, 0, len(c.UseApps))
		for _, init := range c.UseApps {
			apps = append(apps, init())
		}
		registry.Register[Client](NewClient(c.Network, c.Addr, append(c.Options, WithNexusGateway(apps...))...))
	}
}

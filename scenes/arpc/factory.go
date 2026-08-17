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

func (c ClientFactory) Init(container *registry.Container) {
	registry.Export[Client](container, NewClient(c.Network, c.Addr, c.Options...))
}

type NexusClientFactory struct {
	scene.ModuleFactory
	Network string
	Addr    string
	Options []ClientOption
	UseApps []scene.AppInit[ARpcApp]
}

func (c NexusClientFactory) Init(container *registry.Container) {
	apps := make([]ARpcApp, 0, len(c.UseApps))
	for _, init := range c.UseApps {
		app := registry.Load(container, init())
		apps = append(apps, app)
	}
	client := NewClient(c.Network, c.Addr, append(c.Options, WithNexusGateway(apps...))...)
	registry.Export[Client](container, client)
}

package factory

import (
	"github.com/rhine-tech/scene"
	permcmd "github.com/rhine-tech/scene/lens/permission/cmd"
	"github.com/rhine-tech/scene/lens/permission/delivery"
	"github.com/rhine-tech/scene/lens/permission/gen/arpcimpl"
)

type AppGin struct {
	scene.ModuleFactory
}

func (b AppGin) Apps() []scene.Application {
	return []scene.Application{
		delivery.NewGinApp(),
	}
}

type AppARpc struct {
	scene.ModuleFactory
}

func (b AppARpc) Apps() []scene.Application {
	return []scene.Application{
		new(arpcimpl.ARpcAppPermissionService),
	}
}

type AppCmd struct {
	scene.ModuleFactory
}

func (b AppCmd) Apps() []scene.Application {
	return []scene.Application{
		permcmd.NewCmdApp(),
	}
}

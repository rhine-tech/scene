package factory

import (
	"github.com/rhine-tech/scene"
	permcmd "github.com/rhine-tech/scene/lens/permission/cmd"
	"github.com/rhine-tech/scene/lens/permission/delivery"
)

type AppGin struct {
	scene.ModuleFactory
}

func (b AppGin) Apps() []scene.Application {
	return []scene.Application{
		delivery.NewGinApp(),
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

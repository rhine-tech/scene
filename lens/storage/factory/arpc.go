package factory

import (
	"github.com/rhine-tech/scene"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/gen/arpcimpl"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
)

type ServiceARpc struct {
	scene.ModuleFactory
	Client sarpc.Client
}

func (b ServiceARpc) Init() scene.LensInit {
	return func() {
		registry.Register[storageApi.IStorageService](arpcimpl.NewARpcIStorageService(b.Client))
	}
}

type AppARpc struct {
	scene.ModuleFactory
}

func (b AppARpc) Apps() []any {
	return []any{
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](new(arpcimpl.ARpcAppIStorageService))
		},
	}
}

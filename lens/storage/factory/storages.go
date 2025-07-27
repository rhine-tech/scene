package factory

import (
	"github.com/rhine-tech/scene"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/repository/storage"
	"github.com/rhine-tech/scene/registry"
)

type StorageProvider scene.IModuleDependencyProvider[storageApi.IStorageProvider]

type Local struct {
	Root string
}

func (l Local) Default() Local {
	return Local{
		Root: registry.Config.GetString("storage.local.root"),
	}
}

func (l Local) Provide() storageApi.IStorageProvider {
	return registry.Load(storage.NewLocalStorage("default", l.Root))
}

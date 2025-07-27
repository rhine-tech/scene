package factory

import (
	"github.com/rhine-tech/scene"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/delivery"
	"github.com/rhine-tech/scene/lens/storage/repository/meta"
	"github.com/rhine-tech/scene/lens/storage/repository/sessiontracker"
	"github.com/rhine-tech/scene/lens/storage/repository/storage"
	"github.com/rhine-tech/scene/lens/storage/service"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
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

type App struct {
	DefaultProvider string
	Providers       []StorageProvider
}

func (a App) Init() scene.LensInit {
	return func() {
		providers := make([]storageApi.IStorageProvider, 0, len(a.Providers))
		for _, p := range a.Providers {
			providers = append(providers, p.Provide())
		}
		if a.DefaultProvider == "" && len(providers) > 0 {
			a.DefaultProvider = providers[0].ProviderName()
		}
		registry.Register[storageApi.IStorageService](service.NewStorageService(
			registry.Load(meta.NewGormFileMetaRepository()),
			registry.Load(sessiontracker.NewRedisUploadSessionTracker()),
			a.DefaultProvider,
			providers...))
	}
}

func (a App) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return registry.Load(delivery.GinApp())
		},
	}
}

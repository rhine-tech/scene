package factory

import (
	"github.com/rhine-tech/scene"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/delivery"
	"github.com/rhine-tech/scene/lens/storage/repository/meta"
	"github.com/rhine-tech/scene/lens/storage/service"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type Service struct {
	DefaultProvider string
	Providers       []StorageProvider
	SessionTracker  SessionTrackerProvider
}

func (a Service) Default() Service {
	return Service{
		Providers: []StorageProvider{
			Local{}.Default(),
		},
		SessionTracker: SessionTrackerRedis{},
	}
}

func (a Service) Init() scene.LensInit {
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
			a.SessionTracker.Provide(),
			a.DefaultProvider,
			providers...))
	}
}

func (a Service) Apps() []any {
	return []any{}
}

type App struct {
	Service
}

func (a App) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return registry.Load(delivery.GinApp())
		},
	}
}

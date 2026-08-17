package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/delivery"
	"github.com/rhine-tech/scene/lens/storage/repository/meta"
	"github.com/rhine-tech/scene/lens/storage/service"
	"github.com/rhine-tech/scene/registry"
)

type Service struct {
	DefaultProvider string
	Providers       []StorageProvider
	SessionTracker  SessionTrackerProvider
}

func (a Service) Default() Service {
	cfg := registry.Use[config.IConfig](nil)
	providers := []StorageProvider{
		Local{}.Default(),
	}
	s3 := S3{}.Default()
	if s3.Endpoint != "" && s3.Bucket != "" {
		providers = append(providers, s3)
	}
	return Service{
		DefaultProvider: cfg.GetString("storage.default_provider"),
		Providers:       providers,
		SessionTracker:  SessionTrackerMemory{},
	}
}

func (a Service) Init(container *registry.Container) {
	providers := make([]storageApi.IStorageProvider, 0, len(a.Providers))
	for _, providerConfig := range a.Providers {
		provider := registry.Load(container, providerConfig.Provide())
		providers = append(providers, provider)
	}
	if a.DefaultProvider == "" && len(providers) > 0 {
		a.DefaultProvider = providers[0].ProviderName()
	}
	metaRepository := registry.Load(container, meta.NewGormFileMetaRepository(nil))
	uploadSessions := registry.Load(container, a.SessionTracker.Provide())
	registry.Export[storageApi.IStorageService](container, service.NewStorageService(
		metaRepository,
		uploadSessions,
		a.DefaultProvider,
		providers...,
	))
}

func (a Service) Apps() []scene.Application {
	return nil
}

type App struct {
	scene.ModuleFactory
}

func (a App) Apps() []scene.Application {
	return []scene.Application{
		delivery.GinApp(),
	}
}

package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/infrastructure/logger/repository"
	"github.com/rhine-tech/scene/registry"
)

type ZapFactory struct {
	scene.ModuleFactory
	LogLevel logger.LogLevel
	Prefix   string
}

func (b ZapFactory) Default() ZapFactory {
	cfg := registry.Use[config.IConfig](nil)
	return ZapFactory{
		LogLevel: logger.LogLevel(cfg.GetInt("scene.log.level")),
		Prefix:   cfg.GetString("scene.name"),
	}
}

func (b ZapFactory) Init(container *registry.Container) {
	baseLogger := repository.NewZapColoredLogger()
	baseLogger.SetLogLevel(b.LogLevel)
	container.Load(baseLogger)
	log := baseLogger.WithPrefix(b.Prefix)
	registry.Export[logger.ILogger](container, log)
}

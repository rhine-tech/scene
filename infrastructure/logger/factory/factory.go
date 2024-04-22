package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/config"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/infrastructure/logger/repository"
	"github.com/rhine-tech/scene/registry"
)

// Init is instance of scene.LensInit
//func Init() {
//	cfg := registry.AcquireSingleton(config.ConfigUnmarshaler(nil))
//	l := repository.NewLogrusLogger(
//		cfg.GetString("scene.log.file"), cfg.GetInt("scene.log.max_size"),
//		cfg.GetBool("scene.log.panic"),
//	)
//	l.SetLogLevel(logger.LogLevel(cfg.GetInt("scene.log.level")))
//	registry.RegisterLogger(l.WithPrefix(cfg.GetString("scene.name")))
//}

// LogrusFactory
// Deprecated: FunctionName is deprecated.
//type LogrusFactory struct {
//	scene.ModuleFactory
//}
//
//func (b LogrusFactory) Init() scene.LensInit {
//	return Init
//}

type ZapFactory struct {
	scene.ModuleFactory
	LogLevel logger.LogLevel
	Prefix   string
}

func (b ZapFactory) Default() ZapFactory {
	cfg := registry.AcquireSingleton(config.ConfigUnmarshaler(nil))
	return ZapFactory{
		LogLevel: logger.LogLevel(cfg.GetInt("scene.log.level")),
		Prefix:   cfg.GetString("scene.name"),
	}
}

func (b ZapFactory) Init() scene.LensInit {
	return func() {
		l := repository.NewZapColoredLogger()
		l.SetLogLevel(b.LogLevel)
		registry.RegisterLogger(l.WithPrefix(b.Prefix))
	}
}

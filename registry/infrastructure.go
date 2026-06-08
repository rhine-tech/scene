package registry

import (
	"github.com/rhine-tech/scene/infrastructure/asynctask"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/logger"
)

// Infrastructure

var TaskDispatcher asynctask.TaskDispatcher = nil
var Config config.IConfig
var Logger logger.ILogger

func AcquireInfrastructure() {
	TaskDispatcher = AcquireSingleton(asynctask.TaskDispatcher(nil))
	Config = AcquireSingleton(config.IConfig(nil))
	Logger = AcquireSingleton(logger.ILogger(nil))
}

func RegisterConfig(cfg config.IConfig) {
	Config = Register(cfg)
	RegisterSingleton[config.ConfigUnmarshaler](cfg)
	RegisterSingleton[config.ConfigProviderWithDefault](cfg)
}

func RegisterLogger(logger logger.ILogger) {
	Logger = Register(logger)
}

func RegisterTaskDispatcher(taskDispatcher asynctask.TaskDispatcher) {
	TaskDispatcher = Register(taskDispatcher)
}

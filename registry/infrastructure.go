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
	TaskDispatcher = Provide[asynctask.TaskDispatcher]()
	Config = Provide[config.IConfig]()
	Logger = Provide[logger.ILogger]()
}

func RegisterConfig(cfg config.IConfig) {
	Config = Register(cfg)
	Register[config.ConfigUnmarshaler](cfg)
	Register[config.ConfigProviderWithDefault](cfg)
}

func RegisterLogger(logger logger.ILogger) {
	Logger = Register(logger)
}

func RegisterTaskDispatcher(taskDispatcher asynctask.TaskDispatcher) {
	TaskDispatcher = Register(taskDispatcher)
}

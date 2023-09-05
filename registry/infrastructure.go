package registry

import (
	"github.com/rhine-tech/scene/lens/infrastructure/asynctask"
	"github.com/rhine-tech/scene/lens/infrastructure/config"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
)

// Infrastructure

var TaskDispatcher asynctask.TaskDispatcher = nil
var Config config.ConfigUnmarshaler
var Logger logger.ILogger

func AcquireInfrastructure() {
	TaskDispatcher = AcquireSingleton(asynctask.TaskDispatcher(nil))
	Config = AcquireSingleton(config.ConfigUnmarshaler(nil))
	Logger = AcquireSingleton(logger.ILogger(nil))
}

func RegisterConfig(config config.ConfigUnmarshaler) {
	Config = Register(config)
}

func RegisterLogger(logger logger.ILogger) {
	Logger = Register(logger)
}

func RegisterTaskDispatcher(taskDispatcher asynctask.TaskDispatcher) {
	TaskDispatcher = Register(taskDispatcher)
}

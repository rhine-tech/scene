package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
)

const SceneName = "scene.app-container.http.gin"

type GinApplication interface {
	scene.Application
	Prefix() string
	Create(engine *gin.Engine, router gin.IRouter) error
	Destroy() error
}

type CommonApp struct {
	AppError  error
	AppStatus scene.AppStatus
	Logger    logger.ILogger
}

func (s *CommonApp) Status() scene.AppStatus {
	return s.AppStatus
}

func (s *CommonApp) Error() error {
	return s.AppError
}

func (s *CommonApp) Destroy() error {
	return nil
}

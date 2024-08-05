package factory

import (
	"github.com/gin-gonic/gin"
	permMdw "github.com/rhine-tech/scene/lens/permission/delivery/middleware"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func GinWithPermissionContextAuthFacade() sgin.GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(permMdw.GinPermContextAuthFacade(nil))
		return nil
	}
}

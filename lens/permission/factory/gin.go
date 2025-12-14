package factory

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/lens/permission"
	permMdw "github.com/rhine-tech/scene/lens/permission/middleware"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func GinWithPermissionContextFromAuth() sgin.GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(permMdw.GinPermContextFromAuth(registry.Use[permission.PermissionService](nil)))
		return nil
	}
}

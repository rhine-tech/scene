package factory

import (
	"github.com/gin-gonic/gin"
	authMw "github.com/rhine-tech/scene/lens/authentication/delivery/middleware"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func GinWithAuthContext() sgin.GinOption {
	return func(engine *gin.Engine) error {
		engine.Use(authMw.GinAuthContext(nil))
		return nil
	}
}

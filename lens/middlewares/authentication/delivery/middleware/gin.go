package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

func GinAuthContext(verifier authentication.HTTPLoginStatusVerifier) gin.HandlerFunc {
	verifier = registry.Use(verifier)
	infoSrv := registry.Use(authentication.UserInfoService(nil))
	return func(context *gin.Context) {
		ctx := sgin.GetContext(context)
		status, err := verifier.Verify(context.Request)
		if err == nil {
			scene.ContextSetValue(ctx, authentication.NewAuthContext(status.UserID, infoSrv))
		} else {
			scene.ContextSetValue(ctx, authentication.AuthContext{})
		}
		context.Next()
	}
}

// GinRequireAuth is a middleware that requires LoginStatus authentication.
func GinRequireAuth(verifier authentication.HTTPLoginStatusVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		status, err := verifier.Verify(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(err))
			return
		}
		scene.ContextSetValue(ctx, authentication.AuthContext{UserID: status.UserID})
		c.Next()
	}
}

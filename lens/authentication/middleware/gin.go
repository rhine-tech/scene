package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

func GinAuthContext(verifiers ...authentication.HTTPLoginStatusVerifier) gin.HandlerFunc {
	return func(context *gin.Context) {
		ctx := sgin.GetContext(context)
		verified := false
		for _, verifier := range verifiers {
			status, err := verifier.Verify(context.Request)
			if err == nil {
				scene.ContextSetValue(ctx, authentication.NewAuthContext(status.UserID))
				verified = true
				break
			}
		}
		if !verified {
			scene.ContextSetValue(ctx, authentication.AuthContext{})
		}
		context.Next()
	}
}

// GinRequireAuth is a middleware that check if request has been authenticated using authentication.AuthContext
func GinRequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		actx, ok := authentication.GetAuthContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(authentication.ErrNotLogin))
			return
		}
		if !actx.IsLogin() {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(authentication.ErrNotLogin))
		}
		c.Next()
	}
}

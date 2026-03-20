package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
)

func GinAuthContext(verifiers ...authentication.HTTPLoginStatusVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqCtx := c.Request.Context()
		verified := false
		for _, verifier := range verifiers {
			status, err := verifier.Verify(c.Request)
			if err == nil {
				reqCtx = authentication.SetAuthContext(reqCtx, status.UserID)
				verified = true
				break
			}
		}
		if !verified {
			reqCtx = authentication.SetAuthContext(reqCtx, "")
		}
		c.Request = c.Request.WithContext(reqCtx)
		c.Next()
	}
}

// GinRequireAuth is a middleware that check if request has been authenticated using authentication.AuthContext
func GinRequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		actx, ok := authentication.GetAuthContext(c.Request.Context())
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

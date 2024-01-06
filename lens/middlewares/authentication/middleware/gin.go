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

// GinRequireStatusAuth is a middleware that requires LoginStatus authentication.
func GinRequireStatusAuth(lgStSrv authentication.LoginStatusService) gin.HandlerFunc {
	lgStSrv = registry.Use(lgStSrv)
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		status, err := lgStSrv.Verify(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(err))
			return
		}
		scene.ContextSetValue(ctx, authentication.AuthContext{UserID: status.UserID, Username: status.Name})
		c.Next()
	}
}

// GinRequireBasicAuth is a middleware that requires Http BasicAuth authentication.
func GinRequireBasicAuth(authSrv authentication.AuthenticationService) gin.HandlerFunc {
	authSrv = registry.Use(authSrv)
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		user, password, ok := c.Request.BasicAuth()
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(authentication.ErrNotLogin))
			return
		}
		if uid, err := authSrv.Authenticate(user, password); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(err))
			return
		} else {
			scene.ContextSetValue(ctx, authentication.AuthContext{UserID: uid, Username: user})
		}
		c.Next()
	}
}

func GinRequireAnyAuth(useStatus, useBasic bool) gin.HandlerFunc {
	lgStSrv := registry.Use(authentication.LoginStatusService(nil))
	authSrv := registry.Use(authentication.AuthenticationService(nil))
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		if useStatus {
			status, err := lgStSrv.Verify(c.Request)
			if err == nil {
				scene.ContextSetValue(ctx, authentication.AuthContext{UserID: status.UserID, Username: status.Name})
				c.Next()
				return
			}
		}
		if useBasic {
			user, password, ok := c.Request.BasicAuth()
			if ok {
				if uid, err := authSrv.Authenticate(user, password); err == nil {
					c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(err))
					return
				} else {
					scene.ContextSetValue(ctx, authentication.AuthContext{UserID: uid, Username: user})
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(authentication.ErrNotLogin))
		return
	}
}

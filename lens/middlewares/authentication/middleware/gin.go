package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"github.com/rhine-tech/scene/lens/middlewares/permission"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

const ContextKeyStatus = "authentication.status"

func RequirePermUsingAuthGlobal(perm string) func(c *gin.Context) {
	return RequirePermUsingAuth(perm,
		registry.AcquireSingleton((authentication.LoginStatusService)(nil)),
		registry.AcquireSingleton((permission.PermissionService)(nil)))
}

func RequireAuthGlobal() func(c *gin.Context) {
	return RequireAuth(registry.AcquireSingleton((authentication.LoginStatusService)(nil)))
}

func RequireAuth(lgStSrv authentication.LoginStatusService) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		status, err := lgStSrv.Verify(c.Request)
		scene.ContextSetValue(ctx, authentication.AuthContext{UserID: status.UserID, Username: status.Name})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(err))
			return
		}
		c.Next()
	}
}

func RequirePermUsingAuth(
	perm string,
	lgStSrv authentication.LoginStatusService,
	permSrv permission.PermissionService) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		status, err := lgStSrv.Verify(c.Request)
		scene.ContextSetValue(ctx, authentication.AuthContext{UserID: status.UserID, Username: status.Name})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, model.TryErrorCodeResponse(err))
			return
		}
		scene.ContextSetValue(ctx, permission.NewPermContext(permission.PermOwner(status.UserID), permSrv))
		if permSrv.HasPermissionStr(status.UserID, perm) {
			c.Set(ContextKeyStatus, status)
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr(perm)))
	}
}

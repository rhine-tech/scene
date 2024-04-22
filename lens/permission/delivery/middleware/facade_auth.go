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

func GinRequirePermissionAuthFacade(perm string, srv permission.PermissionService) gin.HandlerFunc {
	srv = registry.Use(srv)
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		actx, exist := scene.ContextFindValue[authentication.AuthContext](ctx)
		if !exist || !actx.IsLogin() {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr("owner is empty")))
			return
		}
		scene.ContextSetValue(ctx, permission.NewPermContext(permission.PermOwner(actx.UserID), srv))
		if !srv.HasPermissionStr(actx.UserID, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr(perm)))
			return
		}
		c.Next()
	}
}

func GinPermContextAuthFacade(srv permission.PermissionService) gin.HandlerFunc {
	srv = registry.Use(srv)
	return GinPermContext(func(c *gin.Context) permission.PermOwner {
		ctx := sgin.GetContext(c)
		actx, exist := scene.ContextFindValue[authentication.AuthContext](ctx)
		if !exist || !actx.IsLogin() {
			return ""
		}
		return permission.PermOwner(actx.UserID)
	}, srv)
}

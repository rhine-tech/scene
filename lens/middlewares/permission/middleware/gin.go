package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/lens/middlewares/permission"
	"github.com/rhine-tech/scene/model"
	"net/http"
)

type GinOwnerGetter func(c *gin.Context) permission.PermOwner

func GinRequirePermissionFromRole(srv permission.PermissionService, perm string, getter GinOwnerGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		owner := getter(c)
		if owner == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr("owner is empty")))
			return
		}
		if !srv.HasPermission(string(owner), perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr(perm)))
			return
		}
		c.Next()
	}
}

type GinPermsGetter func(c *gin.Context) permission.PermissionSet

func GinRequirePermission(perm string, getter GinPermsGetter) gin.HandlerFunc {
	requiredPerm := permission.MustParsePermission(perm)
	return func(c *gin.Context) {
		perms := getter(c)
		if perms.HasPermission(requiredPerm) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr(perm)))
	}
}

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	"net/http"
)

func GinPermContextFromAuth(srv permission.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqCtx := c.Request.Context()
		reqCtx = permission.SetPermContext(reqCtx, "", srv)
		actx, exist := authentication.GetAuthContext(reqCtx)
		// if not exists, auth context has not been set properly, so mark it as same as not login
		// for not login, owner will always be empty.
		if !exist || !actx.IsLogin() {
			reqCtx = permission.SetPermContext(reqCtx, "", srv)
		} else {
			reqCtx = permission.SetPermContext(reqCtx, actx.UserID, srv)
		}
		c.Request = c.Request.WithContext(reqCtx)
		c.Next()
	}
}

func GinRequirePermissionFromStr(perm string) gin.HandlerFunc {
	requiredPerm := permission.MustParsePermission(perm)
	return GinRequirePermission(requiredPerm)
}

func GinRequirePermission(requiredPerm *permission.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		permCtx, ok := permission.GetPermContext(c.Request.Context())
		if !ok || !permCtx.HasPermission(requiredPerm) {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr(requiredPerm.String())))
			return
		}
		c.Next()
	}
}

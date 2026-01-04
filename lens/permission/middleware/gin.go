package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

func GinPermContextFromAuth(srv permission.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := sgin.GetContext(c)
		permission.SetPermContext(ctx, "", srv)
		actx, exist := scene.ContextFindValue[authentication.AuthContext](ctx)
		if !exist || !actx.IsLogin() {
			return
		}
		permission.SetPermContext(ctx, actx.UserID, srv)
		c.Next()
	}
}

func GinRequirePermissionFromStr(perm string) gin.HandlerFunc {
	requiredPerm := permission.MustParsePermission(perm)
	return GinRequirePermission(requiredPerm)
}

func GinRequirePermission(requiredPerm *permission.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		permCtx, ok := permission.GetPermContext(sgin.GetContext(c))
		if !ok || !permCtx.HasPermission(requiredPerm) {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorCodeResponse(permission.ErrPermissionDenied.WithDetailStr(requiredPerm.String())))
			return
		}
		c.Next()
	}
}

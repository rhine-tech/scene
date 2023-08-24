package delivery

import (
	"github.com/aynakeya/scene"
	"github.com/aynakeya/scene/errcode"
	"github.com/aynakeya/scene/lens/infrastructure/logger"
	"github.com/aynakeya/scene/lens/middlewares/authentication"
	authMw "github.com/aynakeya/scene/lens/middlewares/authentication/middleware"
	"github.com/aynakeya/scene/lens/middlewares/permission"
	"github.com/aynakeya/scene/model"
	"github.com/gin-gonic/gin"
	"net/http"

	sgin "github.com/aynakeya/scene/scenes/gin"
)

type ginApp struct {
	sgin.CommonApp
	permSrv permission.PermissionService
}

func NewGinApp(logger logger.ILogger, permSrv permission.PermissionService) sgin.GinApplication {
	return &ginApp{
		CommonApp: sgin.CommonApp{
			Logger: logger.WithPrefix("authentication.app.gin"),
		},
		permSrv: permSrv,
	}
}

func (g *ginApp) Name() scene.AppName {
	return "permission.app.gin"
}

func (g *ginApp) Prefix() string {
	return "perms"
}

func (g *ginApp) Create(engine *gin.Engine, router gin.IRouter) error {
	router.GET("/check", authMw.RequireAuthGlobal(), g.handleCheck)
	router.GET("/list", authMw.RequireAuthGlobal(), g.handleList)
	return nil
}

type checkParam struct {
	Perm string `json:"perm" form:"perm" binding:"required"`
}

func (g *ginApp) handleCheck(c *gin.Context) {
	var param checkParam
	if err := c.ShouldBindQuery(&param); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	status := c.MustGet(authMw.ContextKeyStatus).(authentication.LoginStatus)
	c.JSON(200, model.NewDataResponse(gin.H{
		"permission": param.Perm,
		"has":        g.permSrv.HasPermission(status.UserID, param.Perm)}))
}

func (g *ginApp) handleList(c *gin.Context) {
	status := c.MustGet(authMw.ContextKeyStatus).(authentication.LoginStatus)
	c.JSON(200, model.NewDataResponse(g.permSrv.ListPermissions(status.UserID)))
}

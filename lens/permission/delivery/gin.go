package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	authMw "github.com/rhine-tech/scene/lens/middlewares/authentication/delivery/middleware"
	"github.com/rhine-tech/scene/lens/middlewares/permission"
	"github.com/rhine-tech/scene/model"
	"net/http"

	sgin "github.com/rhine-tech/scene/scenes/gin"
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

func (g *ginApp) Name() scene.ImplName {
	return scene.NewAppImplNameNoVer("permission", "gin")
}

func (g *ginApp) Prefix() string {
	return "perms"
}

func (g *ginApp) Create(engine *gin.Engine, router gin.IRouter) error {
	router.GET("/check", authMw.GinRequireAuth(nil), g.handleCheck)
	router.GET("/list", authMw.GinRequireAuth(nil), g.handleList)
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
	actx, _ := scene.ContextFindValue[authentication.AuthContext](sgin.GetContext(c))
	c.JSON(200, model.NewDataResponse(gin.H{
		"permission": param.Perm,
		"has":        g.permSrv.HasPermissionStr(actx.UserID, param.Perm)}))
}

func (g *ginApp) handleList(c *gin.Context) {
	actx, _ := scene.ContextFindValue[authentication.AuthContext](sgin.GetContext(c))
	c.JSON(200, model.NewDataResponse(g.permSrv.ListPermissions(actx.UserID)))
}

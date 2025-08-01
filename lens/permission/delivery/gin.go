package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/authentication"
	authMw "github.com/rhine-tech/scene/lens/authentication/delivery/middleware"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	"net/http"

	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type ginApp struct {
	permSrv permission.PermissionService
}

func NewGinApp(permSrv permission.PermissionService) sgin.GinApplication {
	return &ginApp{
		permSrv: permSrv,
	}
}

func (g *ginApp) Destroy() error {
	return nil
}

func (g *ginApp) Name() scene.ImplName {
	return permission.Lens.ImplName("PermissionDelivery", "gin")
}

func (g *ginApp) Prefix() string {
	return "perms"
}

func (g *ginApp) Create(engine *gin.Engine, router gin.IRouter) error {
	router.GET("/myself/check", authMw.GinRequireAuth(), g.handleCheck)
	router.GET("/myself/list", authMw.GinRequireAuth(), g.handleList)
	router.GET("/list", g.handleAll)
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

func (g *ginApp) handleAll(c *gin.Context) {
	_, ok := authentication.IsLoginInCtx(c)
	if !ok {
		c.JSON(200, model.NewErrorCodeResponse(authentication.ErrNotLogin))
		return
	}
	ok = permission.HasPermissionInCtx(c, permission.PermList)
	if !ok {
		c.JSON(200, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	listType := c.Query("type")
	if listType != "tree" {
		listType = "list"
	}
	if listType == "list" {
		c.JSON(200, model.NewDataResponse(permission.RootPermTree.ToList()))
		return
	}
	c.JSON(200, model.NewDataResponse(permission.RootPermTree.Root.Children))
}

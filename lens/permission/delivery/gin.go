package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/authentication"
	authMw "github.com/rhine-tech/scene/lens/authentication/middleware"
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
	router.GET("/manage/add", authMw.GinRequireAuth(), g.handleManageAdd)
	router.GET("/manage/delete", authMw.GinRequireAuth(), g.handleManageDelete)
	router.GET("/manage/list", authMw.GinRequireAuth(), g.handleManageList)
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
	ctx := sgin.GetContext(c)
	_, ok := authentication.IsLoginInCtx(ctx)
	if !ok {
		c.JSON(200, model.NewErrorCodeResponse(authentication.ErrNotLogin))
		return
	}
	ok = permission.HasPermissionInCtx(ctx, permission.PermList)
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

type managePermParam struct {
	Owner string `json:"owner" form:"owner" binding:"required"`
	Perm  string `json:"perm" form:"perm"`
}

func (g *ginApp) handleManageAdd(c *gin.Context) {
	ctx := sgin.GetContext(c)
	var param managePermParam
	if err := c.ShouldBindQuery(&param); err != nil || param.Perm == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	if _, ok := authentication.IsLoginInCtx(ctx); !ok || !permission.HasPermissionInCtx(ctx, permission.PermManage) {
		c.JSON(http.StatusUnauthorized, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	if err := g.permSrv.AddPermission(param.Owner, param.Perm); err != nil {
		c.JSON(http.StatusOK, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(http.StatusOK, model.NewOkResponse())
}

func (g *ginApp) handleManageDelete(c *gin.Context) {
	ctx := sgin.GetContext(c)
	var param managePermParam
	if err := c.ShouldBindQuery(&param); err != nil || param.Perm == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	if _, ok := authentication.IsLoginInCtx(ctx); !ok || !permission.HasPermissionInCtx(ctx, permission.PermManage) {
		c.JSON(http.StatusUnauthorized, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	if err := g.permSrv.RemovePermission(param.Owner, param.Perm); err != nil {
		c.JSON(http.StatusOK, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(http.StatusOK, model.NewOkResponse())
}

func (g *ginApp) handleManageList(c *gin.Context) {
	ctx := sgin.GetContext(c)
	var param managePermParam
	if err := c.ShouldBindQuery(&param); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	if _, ok := authentication.IsLoginInCtx(ctx); !ok || !permission.HasPermissionInCtx(ctx, permission.PermManage) {
		c.JSON(http.StatusUnauthorized, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	c.JSON(http.StatusOK, model.NewDataResponse(g.permSrv.ListPermissions(param.Owner)))
}

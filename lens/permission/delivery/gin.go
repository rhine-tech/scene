package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/authentication"
	authMw "github.com/rhine-tech/scene/lens/authentication/middleware"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

type ginApp struct {
	permSrv permission.PermissionService `aperture:""`
}

func NewGinApp() sgin.GinApplication {
	return new(ginApp)
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
	router.POST("/manage/add", authMw.GinRequireAuth(), g.handleManageAdd)
	router.DELETE("/manage/delete", authMw.GinRequireAuth(), g.handleManageDelete)
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
	requiredPermission, err := permission.ParsePermission(param.Perm)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	actx, _ := authentication.GetAuthContext(c.Request.Context())
	hasPermission, err := g.permSrv.HasPermission(c.Request.Context(), actx.UserID, requiredPermission)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(200, model.NewDataResponse(gin.H{
		"permission": param.Perm,
		"has":        hasPermission}))
}

func (g *ginApp) handleList(c *gin.Context) {
	actx, _ := authentication.GetAuthContext(c.Request.Context())
	permissions, err := g.permSrv.ListPermissions(c.Request.Context(), actx.UserID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(200, model.NewDataResponse(permissions))
}

func (g *ginApp) handleAll(c *gin.Context) {
	ctx := c.Request.Context()
	_, ok := authentication.IsLoginInCtx(ctx)
	if !ok {
		c.JSON(200, model.NewErrorCodeResponse(authentication.ErrNotLogin))
		return
	}
	ok, err := permission.HasPermissionInCtx(ctx, permission.PermList)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
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
	ctx := c.Request.Context()
	var param managePermParam
	if err := c.ShouldBind(&param); err != nil || param.Perm == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	hasPermission, err := permission.HasPermissionInCtx(ctx, permission.PermManage)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	if _, ok := authentication.IsLoginInCtx(ctx); !ok || !hasPermission {
		c.JSON(http.StatusUnauthorized, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	if err := g.permSrv.AddPermission(ctx, param.Owner, param.Perm); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusOK, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(http.StatusOK, model.NewOkResponse())
}

func (g *ginApp) handleManageDelete(c *gin.Context) {
	ctx := c.Request.Context()
	var param managePermParam
	if err := c.ShouldBind(&param); err != nil || param.Perm == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	hasPermission, err := permission.HasPermissionInCtx(ctx, permission.PermManage)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	if _, ok := authentication.IsLoginInCtx(ctx); !ok || !hasPermission {
		c.JSON(http.StatusUnauthorized, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	if err := g.permSrv.RemovePermission(ctx, param.Owner, param.Perm); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusOK, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(http.StatusOK, model.NewOkResponse())
}

func (g *ginApp) handleManageList(c *gin.Context) {
	ctx := c.Request.Context()
	var param managePermParam
	if err := c.ShouldBindQuery(&param); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(errcode.ParameterError.WithDetail(err)))
		return
	}
	hasPermission, err := permission.HasPermissionInCtx(ctx, permission.PermManage)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	if _, ok := authentication.IsLoginInCtx(ctx); !ok || !hasPermission {
		c.JSON(http.StatusUnauthorized, model.NewErrorCodeResponse(permission.ErrPermissionDenied))
		return
	}
	permissions, err := g.permSrv.ListPermissions(ctx, param.Owner)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, model.NewErrorCodeResponse(errcode.InternalError.WithDetail(err)))
		return
	}
	c.JSON(http.StatusOK, model.NewDataResponse(permissions))
}

package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/middleware"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

type ginApp struct {
	sgin.CommonApp
	lgStSrv authentication.LoginStatusService
	authSrv authentication.AuthenticationService
	infoSrv authentication.UserInfoService
}

func NewGinApp(logger logger.ILogger,
	lgStSrv authentication.LoginStatusService,
	authSrv authentication.AuthenticationService,
	infoSrv authentication.UserInfoService) sgin.GinApplication {
	return &ginApp{
		CommonApp: sgin.CommonApp{
			Logger: logger.WithPrefix("authentication.app.gin"),
		},
		lgStSrv: lgStSrv,
		authSrv: authSrv,
		infoSrv: infoSrv,
	}
}

func (g *ginApp) Create(engine *gin.Engine, router gin.IRouter) error {
	router.GET("/login", g.handleLogin)
	router.POST("/login", g.handleLogin)
	router.GET("/logout", g.handleLogout)
	router.GET("/info", middleware.GinRequireStatusAuth(nil), g.handleInfo)
	//router.POST("/info", middleware.RequireAuthGlobal(), g.handleEditInfo)
	//router.POST("/info/list", middleware.RequireAuthGlobal(), g.handleListInfo)
	//router.POST("/info/delete", middleware.RequireAuthGlobal(), g.handleDeleteInfo)
	//router.POST("/info/update", middleware.RequireAuthGlobal(), g.handleEnableInfo)
	//router.POST("/info/create", middleware.RequireAuthGlobal(), g.handleCreateInfo)
	return nil
}

func (g *ginApp) Name() scene.ImplName {
	return scene.NewAppImplNameNoVer("authentication", "gin")
}

func (g *ginApp) Prefix() string {
	return "auth"
}

type loginParam struct {
	Username string `form:"username" binding:"required" json:"username"`
	Password string `form:"password" binding:"required" json:"password"`
}

func (g *ginApp) handleLogin(c *gin.Context) {
	var param loginParam
	if err := c.ShouldBind(&param); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(
			errcode.ParameterError.WithDetail(err)))
		return
	}
	if userID, err := g.authSrv.Authenticate(param.Username, param.Password); err == nil {
		var status authentication.LoginStatus
		if status, err = g.lgStSrv.Login(userID, c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, model.TryErrorCodeResponse(err))
			return
		}
		c.JSON(http.StatusOK, model.NewDataResponse(gin.H{"username": param.Username, "user_id": userID, "status": status}))
		return
	} else {
		c.JSON(http.StatusOK, model.TryErrorCodeResponse(err))
		return
	}
}

func (g *ginApp) handleLogout(c *gin.Context) {
	err := g.lgStSrv.Logout(c.Writer)
	if err != nil {
		c.JSON(http.StatusOK, model.TryErrorCodeResponse(err))
		return
	}
	c.JSON(http.StatusOK, model.NewOkResponse())
}

func (g *ginApp) handleInfo(c *gin.Context) {
	ctx, _ := scene.ContextFindValue[authentication.AuthContext](sgin.GetContext(c))
	info, err := g.infoSrv.InfoById(ctx.UserID)
	if err != nil {
		c.JSON(http.StatusOK, model.TryErrorCodeResponse(err))
	}
	c.JSON(http.StatusOK, model.NewDataResponse(info))
}

//func (g *ginApp) handleEditInfo(c *gin.Context) {
//	status := c.MustGet(middleware.ContextKeyStatus).(authentication.LoginStatus)
//	var info authentication.UserInfo
//	if err := c.ShouldBind(&info); err != nil {
//		c.JSON(http.StatusBadRequest, model.NewErrorCodeResponse(
//			errcode.ParameterError.WithDetail(err)))
//		return
//	}
//	info.UserID = status.UserID
//	if err := g.infoSrv.EditInfo(info); err != nil {
//		c.JSON(http.StatusOK, model.TryErrorCodeResponse(err))
//		return
//	}
//	c.JSON(http.StatusOK, model.NewOkResponse())
//}

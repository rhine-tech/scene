package delivery

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

type authContext struct {
	authSrv  authentication.IAuthenticationService  `aperture:""`
	tokenSrv authentication.IAccessTokenService     `aperture:""`
	lgStVrf  authentication.HTTPLoginStatusVerifier `aperture:""`
}

// AuthGinApp creates the Gin application definition for all authentication-related routes.
func AuthGinApp(lgStVrf authentication.HTTPLoginStatusVerifier) sgin.GinApplication {
	return &sgin.AppRoutes[authContext]{
		AppName:  authentication.Lens.ImplNameNoVer("GinApplication"),
		BasePath: "auth", // All routes will be prefixed with /auth
		// TODO: Add authentication middleware here for protected routes.
		// For example: Middlewares: []gin.HandlerFunc{middleware.RequireAuthGlobal()},
		Actions: []sgin.Action[*authContext]{
			// User and Session Management
			new(loginRequest),
			new(logoutRequest),
			//new(createUserRequest),
			//new(getUserRequest),
			//new(deleteUserRequest),

			// Access Token (API Key) Management
			new(createTokenRequest),
			new(listTokensRequest),
			new(deleteTokenRequest),
		},
		// The context is initialized empty; DI will populate it.
		Context: authContext{
			lgStVrf: lgStVrf,
		},
	}
}

// --- User and Session Actions ---

// loginRequest handles user login with username and password.
type loginRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

func (l *loginRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{Method: http.MethodGet, Path: "/login", Methods: scene.HttpMethodGet | scene.HttpMethodPost}
}

func (l *loginRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	userID, err := ctx.App.authSrv.Authenticate(l.Username, l.Password)
	if err != nil {
		return nil, err
	}
	status, err := ctx.App.lgStVrf.Login(userID, ctx.Writer)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// logoutRequest handles user logout.
type logoutRequest struct {
	sgin.BaseAction
	sgin.RequestNoParam
}

func (l *logoutRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{Method: http.MethodPost, Path: "/logout"}
}

func (l *logoutRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	return nil, ctx.App.lgStVrf.Logout(ctx.Writer)
}

//// createUserRequest handles the creation of a new user.
//type createUserRequest struct {
//	sgin.BaseAction
//	sgin.RequestQuery
//	Username string `json:"username" binding:"required"`
//	Password string `json:"password" binding:"required"`
//}
//
//func (c *createUserRequest) GetRoute() scene.HttpRouteInfo {
//	return scene.HttpRouteInfo{Method: http.MethodPost, Path: "/users"}
//}
//
//func (c *createUserRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
//	return ctx.App.authSrv.AddUser(c.Username, c.Password)
//}

//// getUserRequest retrieves a user's information by their ID.
//type getUserRequest struct {
//	sgin.BaseAction
//	sgin.RequestURI
//	UserID string `uri:"userId" binding:"required"`
//}
//
//func (g *getUserRequest) GetRoute() scene.HttpRouteInfo {
//	return scene.HttpRouteInfo{Method: http.MethodGet, Path: "/users/:userId"}
//}
//
//func (g *getUserRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
//	// TODO: Add authorization logic here. A user should only be able to get their own info,
//	// unless they are an admin.
//	return ctx.App.authSrv.UserById(g.UserID)
//}
//
//// deleteUserRequest handles deleting a user.
//type deleteUserRequest struct {
//	sgin.BaseAction
//	sgin.RequestURI
//	UserID string `uri:"userId" binding:"required"`
//}
//
//func (d *deleteUserRequest) GetRoute() scene.HttpRouteInfo {
//	return scene.HttpRouteInfo{Method: http.MethodDelete, Path: "/users/:userId"}
//}
//
//func (d *deleteUserRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
//	// TODO: This is a destructive action and requires robust authorization.
//	// An admin should be able to delete any user.
//	// A user might be able to delete their own account.
//	err = ctx.App.authSrv.DeleteUser(d.UserID)
//	if err != nil {
//		return nil, err
//	}
//	return model.NewOkResponse(), nil
//}

// --- Access Token Actions ---

// createTokenRequest creates a new persistent access token (API key).
// This route MUST be protected by authentication middleware.
type createTokenRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Name     string `json:"name" form:"name" binding:"required"`
	UserID   string `json:"user_id" form:"user_id" binding:"required"`
	ExpireAt int64  `json:"expire_at" form:"expire_at,default=-1"` // Unix timestamp, 0 for no expiration
}

func (c *createTokenRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{Method: http.MethodPost, Path: "/token"}
}

func (c *createTokenRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	if c.ExpireAt == 0 {
		c.ExpireAt = -1
	}
	return ctx.App.tokenSrv.WithSceneContext(ctx).Create(c.UserID, c.Name, c.ExpireAt)
}

// listTokensRequest lists all tokens for the currently authenticated user.
// This route MUST be protected by authentication middleware.
type listTokensRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	UserID string `json:"user_id" form:"user_id" binding:"required"`
	Offset int64  `form:"offset,default=0"`
	Limit  int64  `form:"limit,default=20"`
}

func (l *listTokensRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{Method: http.MethodGet, Path: "/tokens"}
}

func (l *listTokensRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	return ctx.App.tokenSrv.WithSceneContext(ctx).ListByUser(l.UserID, l.Offset, l.Limit)
}

// deleteTokenRequest deletes a specific token owned by the user.
// This route MUST be protected by authentication middleware.
type deleteTokenRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Token string `form:"token" binding:"required"`
}

func (d *deleteTokenRequest) GetRoute() scene.HttpRouteInfo {
	return scene.HttpRouteInfo{Method: http.MethodDelete, Path: "/token"}
}

func (d *deleteTokenRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	return nil, ctx.App.tokenSrv.WithSceneContext(ctx).Delete(d.Token)
}

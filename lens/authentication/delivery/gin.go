package delivery

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/service/token"
	"github.com/rhine-tech/scene/lens/permission"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"net/http"
)

type authContext struct {
	authSrv  authentication.IAuthenticationService                 `aperture:""`
	tokenSrv scene.WithContext[authentication.IAccessTokenService] `aperture:"embed"`
	lgStVrf  authentication.HTTPLoginStatusVerifier                `aperture:""`
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
			new(getInfoRequest),
			new(updateProfileRequest),
			new(listUsersRequest),
			new(createUserRequest),
			new(deleteUserRequest),

			// Access Token (API Key) Management
			new(createTokenRequest),
			new(listTokensRequest),
			new(deleteTokenRequest),
		},
		// The context is initialized empty; DI will populate it.
		Context: authContext{
			lgStVrf:  lgStVrf,
			tokenSrv: new(token.CtxProxy),
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

func (l *loginRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodGet, Path: "/login", Methods: sgin.HttpMethodGet | sgin.HttpMethodPost}
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

func (l *logoutRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodGet, Path: "/logout", Methods: sgin.HttpMethodGet | sgin.HttpMethodPost}
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

// getInfoRequest get current user's info
type getInfoRequest struct {
	sgin.BaseAction
	sgin.RequestNoParam
}

func (g *getInfoRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodGet, Path: "/user/info"}
}

func (g *getInfoRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	userId, ok := authentication.IsLoginInCtx(ctx)
	if !ok {
		return nil, authentication.ErrNotLogin
	}
	u, err := ctx.App.authSrv.UserById(userId)
	return UserNoPassword{}.FromUser(u), err
}

// updateProfileRequest allows a logged-in user to update their own profile
type updateProfileRequest struct {
	sgin.BaseAction
	sgin.RequestJson
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
	Timezone    string `json:"timezone"`
	Password    string `json:"password"`
}

func (u *updateProfileRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPut, Path: "/user/profile"}
}

func (u *updateProfileRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	userId, ok := authentication.IsLoginInCtx(ctx)
	if !ok {
		return nil, authentication.ErrNotLogin
	}
	user, err := ctx.App.authSrv.UserById(userId)
	if err != nil {
		return nil, err
	}

	if u.DisplayName != "" {
		user.DisplayName = u.DisplayName
	}
	if u.Email != "" {
		user.Email = u.Email
	}
	if u.Avatar != "" {
		user.Avatar = u.Avatar
	}
	if u.Timezone != "" {
		user.Timezone = u.Timezone
	}
	if u.Password != "" {
		user.Password = u.Password
	}

	if err := ctx.App.authSrv.UpdateUser(user); err != nil {
		return nil, err
	}
	return UserNoPassword{}.FromUser(user), nil
}

// listUsersRequest lists users (admin only)
type listUsersRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Offset int64 `form:"offset,default=0"`
	Limit  int64 `form:"limit,default=20"`
}

func (l *listUsersRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodGet, Path: "/users"}
}

func (l *listUsersRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	if _, ok := authentication.IsLoginInCtx(ctx); !ok {
		return nil, authentication.ErrNotLogin
	}
	if !permission.HasPermissionInCtx(ctx, authentication.PermUserManage) {
		return nil, permission.ErrPermissionDenied
	}
	result, err := ctx.App.authSrv.ListUsers(l.Offset, l.Limit)
	if err != nil {
		return nil, err
	}
	users := make([]UserNoPassword, 0, len(result.Results))
	for _, u := range result.Results {
		users = append(users, UserNoPassword{}.FromUser(u))
	}
	return model.PaginationResult[UserNoPassword]{
		Offset:  result.Offset,
		Results: users,
		Count:   result.Count,
		Total:   result.Total,
	}, nil
}

// createUserRequest handles the creation of a new user.
type createUserRequest struct {
	sgin.BaseAction
	sgin.RequestQuery
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

func (c *createUserRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPost, Path: "/users"}
}

func (c *createUserRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	if _, ok := authentication.IsLoginInCtx(ctx); !ok {
		return nil, authentication.ErrNotLogin
	}
	if !permission.HasPermissionInCtx(ctx, authentication.PermUserManage) {
		return nil, permission.ErrPermissionDenied
	}
	user, err := ctx.App.authSrv.AddUser(c.Username, c.Password)
	if err != nil {
		return nil, err
	}
	return UserNoPassword{}.FromUser(user), nil
}

// deleteUserRequest handles deleting a user.
type deleteUserRequest struct {
	sgin.BaseAction
	sgin.RequestURI
	UserID string `uri:"userId" binding:"required"`
}

func (d *deleteUserRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodDelete, Path: "/users/:userId"}
}

func (d *deleteUserRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	if _, ok := authentication.IsLoginInCtx(ctx); !ok {
		return nil, authentication.ErrNotLogin
	}
	if !permission.HasPermissionInCtx(ctx, authentication.PermUserManage) {
		return nil, permission.ErrPermissionDenied
	}
	err = ctx.App.authSrv.DeleteUser(d.UserID)
	if err != nil {
		return nil, err
	}
	return model.NewOkResponse(), nil
}

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

func (c *createTokenRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPost, Path: "/token"}
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

func (l *listTokensRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodGet, Path: "/tokens"}
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

func (d *deleteTokenRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodDelete, Path: "/token"}
}

func (d *deleteTokenRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	return nil, ctx.App.tokenSrv.WithSceneContext(ctx).Delete(d.Token)
}

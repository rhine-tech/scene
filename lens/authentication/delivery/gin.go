package delivery

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/permission"
	permMdw "github.com/rhine-tech/scene/lens/permission/middleware"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type authContext struct {
	authSrv  authentication.IAuthenticationService `aperture:""`
	tokenSrv authentication.IAccessTokenService    `aperture:""`
	storage  storage.IStorageService               `aperture:""`
	lgStVrf  authentication.HTTPLoginStatusVerifier
}

func hasUserManagePermission(ctx *sgin.Context[*authContext]) bool {
	return permission.HasPermissionInCtx(ctx, authentication.PermUserManage) ||
		permission.HasPermissionInCtx(ctx, authentication.PermAdmin)
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
			new(uploadAvatarRequest),
			new(listUsersRequest),
			new(createUserRequest),
			new(updateUserRequest),
			new(deleteUserRequest),

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
	sgin.RequestNoParam
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

func (l *loginRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPost, Path: "/login"}
}

func (l *loginRequest) Bind(ctx *sgin.Context[*authContext]) error {
	return ctx.ShouldBind(l)
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
	return UserNoPasswordFromUser(u, ctx.App.storage), err
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
	return UserNoPasswordFromUser(user, ctx.App.storage), nil
}

type uploadAvatarRequest struct {
	sgin.BaseAction
	sgin.RequestNoParam
	fileName    string
	contentType string
	content     []byte
}

func (u *uploadAvatarRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPost, Path: "/user/avatar"}
}

func (u *uploadAvatarRequest) Bind(ctx *sgin.Context[*authContext]) error {
	if err := ctx.Request.ParseMultipartForm(6 << 20); err != nil {
		return err
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	limited := &io.LimitedReader{R: file, N: 5 << 20}
	content, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if limited.N <= 0 {
		return errcode.ParameterError.WithDetailStr("avatar too large")
	}
	detectedType := http.DetectContentType(content)
	if !isAllowedAvatarType(detectedType) {
		return errcode.ParameterError.WithDetailStr("unsupported avatar type")
	}
	u.fileName = header.Filename
	u.contentType = detectedType
	u.content = content
	return nil
}

func (u *uploadAvatarRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	userId, ok := authentication.IsLoginInCtx(ctx)
	if !ok {
		return nil, authentication.ErrNotLogin
	}
	user, err := ctx.App.authSrv.UserById(userId)
	if err != nil {
		return nil, err
	}
	if ctx.App.storage == nil {
		return nil, errcode.InternalError.WithDetailStr("storage service not available")
	}
	fileId, err := ctx.App.storage.Store(bytes.NewReader(u.content), storage.FileMeta{
		OriginalFilename: u.fileName,
		ContentType:      u.contentType,
		ContentLength:    int64(len(u.content)),
	})
	if err != nil {
		return nil, err
	}
	user.Avatar = string(fileId)
	if err := ctx.App.authSrv.UpdateUser(user); err != nil {
		return nil, err
	}
	return UserNoPasswordFromUser(user, ctx.App.storage), nil
}

func isAllowedAvatarType(contentType string) bool {
	switch contentType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	default:
		return false
	}
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
	if !hasUserManagePermission(ctx) {
		return nil, permission.ErrPermissionDenied
	}
	result, err := ctx.App.authSrv.ListUsers(l.Offset, l.Limit)
	if err != nil {
		return nil, err
	}
	users := make([]UserNoPassword, 0, len(result.Results))
	for _, u := range result.Results {
		users = append(users, UserNoPasswordFromUser(u, ctx.App.storage))
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
	if !hasUserManagePermission(ctx) {
		return nil, permission.ErrPermissionDenied
	}
	user, err := ctx.App.authSrv.AddUser(c.Username, c.Password)
	if err != nil {
		return nil, err
	}
	return UserNoPasswordFromUser(user, ctx.App.storage), nil
}

// updateUserRequest allows admins to update another user's profile.
type updateUserRequest struct {
	sgin.BaseAction
	sgin.RequestJson
	UserID      string  `uri:"userId" binding:"required"`
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
	Avatar      *string `json:"avatar"`
	Timezone    *string `json:"timezone"`
	Password    *string `json:"password"`
}

func (u *updateUserRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPut, Path: "/users/:userId"}
}

func (u *updateUserRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	if _, ok := authentication.IsLoginInCtx(ctx); !ok {
		return nil, authentication.ErrNotLogin
	}
	if !hasUserManagePermission(ctx) {
		return nil, permission.ErrPermissionDenied
	}

	user, err := ctx.App.authSrv.UserById(ctx.Param("userId"))
	if err != nil {
		return nil, err
	}

	if u.Username != nil {
		next := strings.TrimSpace(*u.Username)
		if next != "" {
			user.Username = next
		}
	}
	if u.DisplayName != nil {
		user.DisplayName = *u.DisplayName
	}
	if u.Email != nil {
		user.Email = *u.Email
	}
	if u.Avatar != nil {
		user.Avatar = *u.Avatar
	}
	if u.Timezone != nil {
		user.Timezone = *u.Timezone
	}
	if u.Password != nil {
		user.Password = *u.Password
	}

	if err := ctx.App.authSrv.UpdateUser(user); err != nil {
		return nil, err
	}
	return UserNoPasswordFromUser(user, ctx.App.storage), nil
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
	if !hasUserManagePermission(ctx) {
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
	sgin.RequestJson
	Name     string `json:"name" form:"name" binding:"required"`
	UserID   string `json:"user_id" form:"user_id" binding:"required"`
	ExpireAt int64  `json:"expire_at" form:"expire_at,default=-1"` // Unix timestamp, 0 for no expiration
}

func (c *createTokenRequest) GetRoute() sgin.HttpRouteInfo {
	return sgin.HttpRouteInfo{Method: http.MethodPost, Path: "/token"}
}

func (c *createTokenRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(authentication.PermTokenCreate),
	}
}

func (c *createTokenRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	if c.ExpireAt == 0 {
		c.ExpireAt = -1
	}
	return ctx.App.tokenSrv.Create(c.UserID, c.Name, c.ExpireAt)
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

func (l *listTokensRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(authentication.PermTokenList),
	}
}

func (l *listTokensRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	return ctx.App.tokenSrv.ListByUser(l.UserID, l.Offset, l.Limit)
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

func (d *deleteTokenRequest) Middleware() gin.HandlersChain {
	return gin.HandlersChain{
		permMdw.GinRequirePermission(authentication.PermTokenDelete),
	}
}

func (d *deleteTokenRequest) Process(ctx *sgin.Context[*authContext]) (data any, err error) {
	return nil, ctx.App.tokenSrv.Delete(d.Token)
}

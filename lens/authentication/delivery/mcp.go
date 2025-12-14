package delivery

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	smcp "github.com/rhine-tech/scene/scenes/mcp"
)

type app struct {
	srv authentication.IAuthenticationService `aperture:""`
}

func NewMcpApp() smcp.McpApp {
	return &app{}
}

func (a *app) Name() scene.ImplName {
	return authentication.Lens.ImplNameNoVer("McpApp")
}

func (a *app) Register(server *server.MCPServer) error {
	server.AddTools(
		//a.toolLoginByToken(),
		a.toolGetMyInfo(),
	)
	return nil
}

func (a *app) toolLoginByToken() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			a.srv.SrvImplName().MethodName("LoginByToken"),
			mcp.WithDescription("使用 access token 进行登录并返回用户信息"),
			mcp.WithString("token", mcp.Required(), mcp.Description("scene access token")),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			token, err := req.RequireString("token")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			userID, err := a.srv.AuthenticateByToken(token)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			return mcp.NewToolResultJSON(map[string]any{
				"user_id":     userID,
				"token":       token,
				"instruction": "登陆成功，对于任何需要鉴权的接口，请求时带上该token",
			})
		},
	}
}

func (a *app) toolGetMyInfo() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
			a.srv.SrvImplName().MethodName("GetMyInfo"),
			mcp.WithDescription("获取当前登录用户信息"),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, sceneCtx := smcp.GetContext(ctx)
			authCtx, ok := authentication.GetAuthContext(sceneCtx)
			if !ok || !authCtx.IsLogin() {
				return mcp.NewToolResultError(authentication.ErrNotLogin.Error()), nil
			}
			user, err := a.srv.UserById(authCtx.UserID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultJSON(UserNoPassword{}.FromUser(user))
		},
	}
}

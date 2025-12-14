package middleware

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene/lens/authentication"
	smcp "github.com/rhine-tech/scene/scenes/mcp"
	"net/http"
	"strings"
)

//// McpAuthContextWithToken configures the MCP server to require a token on tool calls
//// and populates AuthContext for downstream handlers.
//func McpAuthContextWithToken(srv authentication.IAuthenticationService) server.ToolHandlerMiddleware {
//	srv = registry.Use(srv)
//	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
//		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//			ctx, sceneCtx := smcp.GetContext(ctx)
//			authentication.SetAuthContext(sceneCtx, "")
//			token := strings.TrimSpace(mcp.ParseString(req, "scene_token", ""))
//			if token == "" {
//				token = strings.TrimSpace(mcp.ParseString(req, "token", ""))
//			}
//			if token == "" {
//				return next(ctx, req)
//			}
//			userID, err := srv.AuthenticateByToken(token)
//			if err != nil {
//				return next(ctx, req)
//			}
//			authentication.SetAuthContext(sceneCtx, userID)
//			return next(ctx, req)
//		}
//	}
//}

func McpAuthContextFromSSEHeader(
	key string,
	srv authentication.IAuthenticationService,
) smcp.SSEOption {
	if key == "" {
		key = "scene_token"
	}
	return smcp.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
		token := strings.TrimSpace(r.Header.Get(key))
		ctx, sceneCtx := smcp.GetContext(ctx)
		if token == "" {
			return ctx
		}
		userID, err := srv.AuthenticateByToken(token)
		if err != nil {
			return ctx
		}
		authentication.SetAuthContext(sceneCtx, userID)
		return ctx
	})
}

func McpRequireLogin(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_, sceneCtx := smcp.GetContext(ctx)
		authCtx, ok := authentication.GetAuthContext(sceneCtx)
		if !ok {
			return mcp.NewToolResultError(authentication.ErrNotLogin.Error()), nil
		}
		if !authCtx.IsLogin() {
			return mcp.NewToolResultError(authentication.ErrNotLogin.Error()), nil
		}
		return next(ctx, req)
	}
}

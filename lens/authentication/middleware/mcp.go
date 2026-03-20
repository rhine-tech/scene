package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene/lens/authentication"
	smcp "github.com/rhine-tech/scene/scenes/mcp"
)

func McpAuthContextFromSSEHeader(
	key string,
	srv authentication.IAuthenticationService,
) smcp.SSEOption {
	if key == "" {
		key = "scene_token"
	}
	return smcp.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
		token := strings.TrimSpace(r.Header.Get(key))
		if token == "" {
			return authentication.SetAuthContext(ctx, "")
		}
		userID, err := srv.AuthenticateByToken(token)
		if err != nil {
			return authentication.SetAuthContext(ctx, "")
		}
		return authentication.SetAuthContext(ctx, userID)
	})
}

func McpRequireLogin(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authCtx, ok := authentication.GetAuthContext(ctx)
		if !ok {
			return mcp.NewToolResultError(authentication.ErrNotLogin.Error()), nil
		}
		if !authCtx.IsLogin() {
			return mcp.NewToolResultError(authentication.ErrNotLogin.Error()), nil
		}
		return next(ctx, req)
	}
}

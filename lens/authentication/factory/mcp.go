package factory

import (
	"github.com/rhine-tech/scene/lens/authentication"
	authMw "github.com/rhine-tech/scene/lens/authentication/middleware"
	"github.com/rhine-tech/scene/registry"
	smcp "github.com/rhine-tech/scene/scenes/mcp"
)

func McpRequireLogin() smcp.ServerOption {
	return smcp.WithToolHandlerMiddleware(authMw.McpRequireLogin)
}

// McpAuthContextFromSSEHeader extracts scene_token/Bearer from HTTP headers (SSE message endpoint)
// and injects AuthContext before tool handlers run.
func McpAuthContextFromSSEHeader() smcp.SSEOption {
	return authMw.McpAuthContextFromSSEHeader("scene_token", registry.Use[authentication.IAuthenticationService](nil))
}

package mcp

import "github.com/mark3labs/mcp-go/server"

type ServerOption = server.ServerOption
type SSEOption = server.SSEOption

func WithToolHandlerMiddleware(mw server.ToolHandlerMiddleware) server.ServerOption {
	return server.WithToolHandlerMiddleware(mw)
}

func WithResourceHandlerMiddleware(mw server.ResourceHandlerMiddleware) server.ServerOption {
	return server.WithResourceHandlerMiddleware(mw)
}

func WithToolFilter(filter server.ToolFilterFunc) server.ServerOption {
	return server.WithToolFilter(filter)
}

func WithSSEContextFunc(fn server.SSEContextFunc) server.SSEOption {
	return server.WithSSEContextFunc(fn)
}

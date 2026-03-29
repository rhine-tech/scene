package mcp

import "github.com/mark3labs/mcp-go/server"

type ServerOption = server.ServerOption
type SSEOption = server.SSEOption
type StreamableHTTPOption = server.StreamableHTTPOption

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

func WithStreamableHTTPEndpointPath(path string) server.StreamableHTTPOption {
	return server.WithEndpointPath(path)
}

func WithStreamableHTTPContextFunc(fn server.HTTPContextFunc) server.StreamableHTTPOption {
	return server.WithHTTPContextFunc(fn)
}

func WithStreamableHTTPStateful(stateful bool) server.StreamableHTTPOption {
	return server.WithStateful(stateful)
}

func WithStreamableHTTPStateless(stateless bool) server.StreamableHTTPOption {
	return server.WithStateLess(stateless)
}

func WithStreamableHTTPDisableStreaming(disable bool) server.StreamableHTTPOption {
	return server.WithDisableStreaming(disable)
}

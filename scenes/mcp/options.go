package mcp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

type ServerOption = server.ServerOption

// HTTPServerBuilder creates the HTTP server that owns handler.
type HTTPServerBuilder func(handler http.Handler) *http.Server

// SSEOption configures an SSE handler or its Scene-owned HTTP server.
type SSEOption func(*sseOptions)

type sseOptions struct {
	handlerOptions    []server.SSEOption
	httpServerBuilder HTTPServerBuilder
}

// StreamableHTTPOption configures a Streamable HTTP handler or transport.
type StreamableHTTPOption func(*streamableHTTPOptions)

type streamableHTTPOptions struct {
	handlerOptions    []server.StreamableHTTPOption
	httpServerBuilder HTTPServerBuilder
	endpointPath      string
	tlsCertFile       string
	tlsKeyFile        string
}

func resolveSSEOptions(options []SSEOption) sseOptions {
	resolved := sseOptions{httpServerBuilder: defaultHTTPServerBuilder}
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}
	return resolved
}

func resolveStreamableHTTPOptions(options []StreamableHTTPOption) streamableHTTPOptions {
	resolved := streamableHTTPOptions{
		httpServerBuilder: defaultHTTPServerBuilder,
		endpointPath:      "/mcp",
	}
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}
	return resolved
}

func defaultHTTPServerBuilder(handler http.Handler) *http.Server {
	return &http.Server{Handler: handler}
}

func buildHTTPServer(addr string, handler http.Handler, builder HTTPServerBuilder) (*http.Server, error) {
	if builder == nil {
		return nil, fmt.Errorf("scene mcp: HTTP server builder is nil")
	}
	httpServer := builder(handler)
	if httpServer == nil {
		return nil, fmt.Errorf("scene mcp: HTTP server builder returned nil")
	}
	if httpServer.Addr == "" {
		httpServer.Addr = addr
	} else if httpServer.Addr != addr {
		return nil, fmt.Errorf("scene mcp: HTTP server address %q conflicts with scene address %q", httpServer.Addr, addr)
	}
	if httpServer.Handler == nil {
		httpServer.Handler = handler
	}
	return httpServer, nil
}

func WithToolHandlerMiddleware(mw server.ToolHandlerMiddleware) server.ServerOption {
	return server.WithToolHandlerMiddleware(mw)
}

func WithResourceHandlerMiddleware(mw server.ResourceHandlerMiddleware) server.ServerOption {
	return server.WithResourceHandlerMiddleware(mw)
}

func WithToolFilter(filter server.ToolFilterFunc) server.ServerOption {
	return server.WithToolFilter(filter)
}

// WithSSEHandlerOption applies an mcp-go option to the SSE handler.
// HTTP server ownership remains with the Scene.
func WithSSEHandlerOption(option server.SSEOption) SSEOption {
	return func(options *sseOptions) {
		options.handlerOptions = append(options.handlerOptions, option)
	}
}

func WithSSEContextFunc(fn server.SSEContextFunc) SSEOption {
	return WithSSEHandlerOption(server.WithSSEContextFunc(fn))
}

// WithSSEHTTPServer configures the HTTP server around the SSE handler.
func WithSSEHTTPServer(builder HTTPServerBuilder) SSEOption {
	return func(options *sseOptions) {
		options.httpServerBuilder = builder
	}
}

// WithStreamableHTTPHandlerOption applies an mcp-go option to the Streamable
// HTTP handler. Server, TLS, and endpoint options must use the Scene options.
func WithStreamableHTTPHandlerOption(option server.StreamableHTTPOption) StreamableHTTPOption {
	return func(options *streamableHTTPOptions) {
		options.handlerOptions = append(options.handlerOptions, option)
	}
}

func WithStreamableHTTPEndpointPath(endpointPath string) StreamableHTTPOption {
	return func(options *streamableHTTPOptions) {
		options.endpointPath = "/" + strings.Trim(endpointPath, "/")
	}
}

func WithStreamableHTTPContextFunc(fn server.HTTPContextFunc) StreamableHTTPOption {
	return WithStreamableHTTPHandlerOption(server.WithHTTPContextFunc(fn))
}

func WithStreamableHTTPStateful(stateful bool) StreamableHTTPOption {
	return WithStreamableHTTPHandlerOption(server.WithStateful(stateful))
}

func WithStreamableHTTPStateless(stateless bool) StreamableHTTPOption {
	return WithStreamableHTTPHandlerOption(server.WithStateLess(stateless))
}

func WithStreamableHTTPDisableStreaming(disable bool) StreamableHTTPOption {
	return WithStreamableHTTPHandlerOption(server.WithDisableStreaming(disable))
}

// WithStreamableHTTPServer configures the HTTP server around the routed MCP handler.
func WithStreamableHTTPServer(builder HTTPServerBuilder) StreamableHTTPOption {
	return func(options *streamableHTTPOptions) {
		options.httpServerBuilder = builder
	}
}

// WithStreamableHTTPTLS configures the certificate used by the Scene listener.
func WithStreamableHTTPTLS(certFile, keyFile string) StreamableHTTPOption {
	return func(options *streamableHTTPOptions) {
		options.tlsCertFile = certFile
		options.tlsKeyFile = keyFile
	}
}

func routeStreamableHTTP(endpointPath string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		path := request.URL.Path
		if endpointPath == "/" || path == endpointPath || path == server.WellKnownProtectedResourcePath ||
			strings.HasPrefix(path, server.WellKnownProtectedResourcePath+"/") {
			handler.ServeHTTP(writer, request)
			return
		}
		http.NotFound(writer, request)
	})
}

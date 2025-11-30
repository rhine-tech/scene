package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/rhine-tech/scene"
)

type McpApp interface {
	scene.Application
	Register(server *server.MCPServer) error
}

package factory

import (
	"github.com/rhine-tech/scene"
	authcmd "github.com/rhine-tech/scene/lens/authentication/cmd"
	"github.com/rhine-tech/scene/lens/authentication/delivery"
	"github.com/rhine-tech/scene/lens/authentication/gen/arpcimpl"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	scmd "github.com/rhine-tech/scene/scenes/cmd"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	smcp "github.com/rhine-tech/scene/scenes/mcp"
)

type AppGin struct {
	scene.ModuleFactory
	Verifier HttpVerifier
}

func (b AppGin) Default() AppGin {
	return AppGin{
		Verifier: JWTVerifier{}.Default(),
	}
}

func (b AppGin) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.AuthGinApp(b.Verifier.Provide())
		},
	}
}

type AppArpc struct {
	scene.ModuleFactory
}

func (b AppArpc) Apps() []any {
	return []any{
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](&arpcimpl.ARpcAppIAuthenticationService{})
		},
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](&arpcimpl.ARpcAppIAccessTokenService{})
		},
	}
}

type AppMcp struct {
	scene.ModuleFactory
}

func (b AppMcp) Apps() []any {
	return []any{
		func() smcp.McpApp {
			return registry.Load(delivery.NewMcpApp())
		},
	}
}

type AppCmd struct {
	scene.ModuleFactory
}

func (b AppCmd) Apps() []any {
	return []any{
		func() scmd.CmdApp {
			return registry.Load(authcmd.NewCmdApp())
		},
	}
}

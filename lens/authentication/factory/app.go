package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication/delivery"
	"github.com/rhine-tech/scene/lens/authentication/gen/arpcimpl"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	sgin "github.com/rhine-tech/scene/scenes/gin"
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

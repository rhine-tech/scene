package factory

import (
	"github.com/rhine-tech/scene"
	authcmd "github.com/rhine-tech/scene/lens/authentication/cmd"
	"github.com/rhine-tech/scene/lens/authentication/delivery"
	"github.com/rhine-tech/scene/lens/authentication/gen/arpcimpl"
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

func (b AppGin) Apps() []scene.Application {
	return []scene.Application{
		delivery.AuthGinApp(b.Verifier.Provide()),
	}
}

type AppArpc struct {
	scene.ModuleFactory
}

func (b AppArpc) Apps() []scene.Application {
	return []scene.Application{
		new(arpcimpl.ARpcAppIAuthenticationService),
		new(arpcimpl.ARpcAppIAccessTokenService),
	}
}

type AppMcp struct {
	scene.ModuleFactory
}

func (b AppMcp) Apps() []scene.Application {
	return []scene.Application{
		delivery.NewMcpApp(),
	}
}

type AppCmd struct {
	scene.ModuleFactory
}

func (b AppCmd) Apps() []scene.Application {
	return []scene.Application{
		authcmd.NewCmdApp(),
	}
}

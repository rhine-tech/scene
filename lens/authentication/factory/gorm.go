package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/delivery"
	"github.com/rhine-tech/scene/lens/authentication/repository"
	"github.com/rhine-tech/scene/lens/authentication/service"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type GinVerifier scene.IModuleDependencyProvider[authentication.HTTPLoginStatusVerifier]

type GinAppGorm struct {
	scene.ModuleFactory
	Verifier GinVerifier
}

func (b GinAppGorm) Default() GinAppGorm {
	return GinAppGorm{
		Verifier: JWTVerifier{}.Default(),
	}
}

func (b GinAppGorm) Init() scene.LensInit {
	return func() {
		repo := registry.Load(repository.NewGormAuthenticationRepository(nil))
		repo2 := registry.Load(repository.NewGormAccessTokenRepository(nil))
		srv := registry.Register[authentication.IAccessTokenService](service.NewAccessTokenService(repo2, nil))
		registry.Register[authentication.IAuthenticationService](service.NewAuthenticationService(nil, repo, srv))
		registry.Register[authentication.HTTPLoginStatusVerifier](b.Verifier.Provide())
	}
}

func (b GinAppGorm) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.AuthGinApp()
		},
	}
}

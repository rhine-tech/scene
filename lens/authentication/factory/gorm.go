package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/delivery"
	"github.com/rhine-tech/scene/lens/authentication/gen/arpcimpl"
	"github.com/rhine-tech/scene/lens/authentication/repository"
	"github.com/rhine-tech/scene/lens/authentication/service"
	"github.com/rhine-tech/scene/lens/authentication/service/token"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type AppGorm struct {
	scene.ModuleFactory
	Verifier HttpVerifier
}

func (b AppGorm) Default() AppGorm {
	return AppGorm{
		Verifier: JWTVerifier{}.Default(),
	}
}

func (b AppGorm) Init() scene.LensInit {
	return func() {
		repo := registry.Load(repository.NewGormAuthenticationRepository(nil))
		repo2 := registry.Load(repository.NewGormAccessTokenRepository(nil))
		srv := registry.Register[authentication.IAccessTokenService](token.NewAccessTokenService(repo2, nil))
		registry.Register[authentication.IAuthenticationService](service.NewAuthenticationService(nil, repo, srv))
	}
}

func (b AppGorm) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.AuthGinApp(b.Verifier.Provide())
		},
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](&arpcimpl.ARpcAppIAuthenticationService{})
		},
		func() sarpc.ARpcApp {
			return registry.Load[sarpc.ARpcApp](&arpcimpl.ARpcAppIAccessTokenService{})
		},
	}
}

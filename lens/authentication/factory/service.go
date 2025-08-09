package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/gen/arpcimpl"
	"github.com/rhine-tech/scene/lens/authentication/repository"
	"github.com/rhine-tech/scene/lens/authentication/service"
	"github.com/rhine-tech/scene/lens/authentication/service/token"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
)

type ServiceARpc struct {
	scene.ModuleFactory
	Client sarpc.Client
}

func (b ServiceARpc) Init() scene.LensInit {
	return func() {
		registry.Register[authentication.IAccessTokenService](arpcimpl.NewARpcIAccessTokenService(b.Client))
		registry.Register[authentication.IAuthenticationService](arpcimpl.NewARpcIAuthenticationService(b.Client))
	}
}

type ServiceGorm struct {
	scene.ModuleFactory
}

func (b ServiceGorm) Init() scene.LensInit {
	return func() {
		repo := registry.Load(repository.NewGormAuthenticationRepository(nil))
		repo2 := registry.Load(repository.NewGormAccessTokenRepository(nil))
		srv := registry.Register[authentication.IAccessTokenService](token.NewAccessTokenService(repo2, nil))
		registry.Register[authentication.IAuthenticationService](service.NewAuthenticationService(nil, repo, srv))
	}
}

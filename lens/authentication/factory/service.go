package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/gen/arpcimpl"
	"github.com/rhine-tech/scene/lens/authentication/repository"
	"github.com/rhine-tech/scene/lens/authentication/service/base"
	"github.com/rhine-tech/scene/lens/authentication/service/proxy"
	"github.com/rhine-tech/scene/lens/authentication/service/token"
	"github.com/rhine-tech/scene/registry"
	sarpc "github.com/rhine-tech/scene/scenes/arpc"
)

type ServiceARpc struct {
	scene.ModuleFactory
	Client sarpc.Client
}

func (b ServiceARpc) Init(container *registry.Container) {
	registry.Export[authentication.IAccessTokenService](container, arpcimpl.NewARpcIAccessTokenService(b.Client))
	registry.Export[authentication.IAuthenticationService](container, arpcimpl.NewARpcIAuthenticationService(b.Client))
}

type ServiceGorm struct {
	scene.ModuleFactory
}

func (b ServiceGorm) Init(container *registry.Container) {
	authenticationRepository := registry.Load(container, repository.NewGormAuthenticationRepository(nil))
	tokenRepository := registry.Load(container, repository.NewGormAccessTokenRepository(nil))
	tokenService := registry.Export[authentication.IAccessTokenService](container, token.NewAccessTokenService(tokenRepository, nil))
	baseService := registry.Load(container, base.NewAuthenticationService(nil, authenticationRepository, tokenService))
	registry.Export[authentication.IAuthenticationService](container, proxy.NewCachedAuthenticationService(baseService))
}

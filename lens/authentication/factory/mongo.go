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

type GinAppMongoDB struct {
	scene.ModuleFactory
	Verifier scene.IModuleDependencyProvider[authentication.HTTPLoginStatusVerifier]
}

func (b GinAppMongoDB) Default() GinAppMongoDB {
	return GinAppMongoDB{
		Verifier: JWTVerifier{
			Key:    "scene_token",
			Secret: []byte(registry.Config.GetString("authentication.jwt.secret")),
		},
	}
}

func (b GinAppMongoDB) Init() scene.LensInit {
	return func() {
		repo := registry.Load(repository.NewMongoAuthenticationRepository(nil))
		repo2 := registry.Load(repository.NewUserInfoRepository(nil))
		srv1 := registry.Register(service.NewAuthenticationService(nil, repo))
		registry.Register[authentication.UserInfoService](service.NewUserInfoService(repo, repo2))
		registry.Register[authentication.AuthenticationService](srv1.(authentication.AuthenticationService))
		registry.Register[authentication.HTTPLoginStatusVerifier](b.Verifier.Provide())
	}
}

func (b GinAppMongoDB) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.NewGinApp(
				registry.Use[authentication.HTTPLoginStatusVerifier](nil),
				nil,
				nil)
		},
	}
}

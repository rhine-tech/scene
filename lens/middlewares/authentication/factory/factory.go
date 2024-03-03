package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/delivery"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/repository"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/service"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/service/loginstatus"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type HTTPLoginStatusVerifier scene.IModuleDependencyProvider[authentication.HTTPLoginStatusVerifier]

type JWTVerifier struct {
	Key    string
	Secret []byte
}

func (J JWTVerifier) Provide() authentication.HTTPLoginStatusVerifier {
	return loginstatus.NewJWT(J.Secret, J.Key)
}

type GinAppMongoDB struct {
	scene.ModuleFactory
	Verifier HTTPLoginStatusVerifier
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
	}
}

func (b GinAppMongoDB) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.NewGinApp(
				registry.Use[logger.ILogger](nil),
				registry.Load(b.Verifier.Provide()),
				nil,
				nil)
		},
	}
}

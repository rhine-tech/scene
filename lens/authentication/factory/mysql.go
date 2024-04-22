package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/lens/authentication/delivery"
	"github.com/rhine-tech/scene/lens/authentication/repository/mysql"
	"github.com/rhine-tech/scene/lens/authentication/service"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type GinAppMysql struct {
	scene.ModuleFactory
	Verifier scene.IModuleDependencyProvider[authentication.HTTPLoginStatusVerifier]
}

func (b GinAppMysql) Default() GinAppMysql {
	return GinAppMysql{
		Verifier: JWTVerifier{
			Key:    "scene_token",
			Secret: []byte(registry.Config.GetString("authentication.jwt.secret")),
		},
	}
}

func (b GinAppMysql) Init() scene.LensInit {
	return func() {
		repo := registry.Load(mysql.AuthenticationRepository(nil))
		repo2 := registry.Load(mysql.NewUserInfoRepository(nil))
		srv1 := registry.Register(service.NewAuthenticationService(nil, repo))
		registry.Register[authentication.UserInfoService](service.NewUserInfoService(repo, repo2))
		registry.Register[authentication.AuthenticationService](srv1.(authentication.AuthenticationService))
		registry.Register[authentication.HTTPLoginStatusVerifier](b.Verifier.Provide())
	}
}

func (b GinAppMysql) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.NewGinApp(
				registry.Use[authentication.HTTPLoginStatusVerifier](nil),
				nil,
				nil)
		},
	}
}

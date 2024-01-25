package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/delivery"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/repository"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/service"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

func CreateApp(logger logger.ILogger,
	authSrv authentication.AuthenticationService,
	lgstSrv authentication.LoginStatusService,
	infoSrv authentication.UserInfoService) sgin.GinApplication {
	return delivery.NewGinApp(logger, lgstSrv, authSrv, infoSrv)
}

func InitApp() sgin.GinApplication {
	return CreateApp(
		registry.AcquireSingleton(logger.ILogger(nil)),
		registry.AcquireSingleton(authentication.AuthenticationService(nil)),
		registry.AcquireSingleton(authentication.LoginStatusService(nil)),
		registry.AcquireSingleton(authentication.UserInfoService(nil)))
}

type AppMongoDB struct {
	scene.ModuleFactory
}

func (b AppMongoDB) Init() scene.LensInit {
	return func() {
		repo := registry.Load(repository.NewMongoAuthenticationRepository(nil))
		repo2 := registry.Load(repository.NewUserInfoRepository(nil))
		srv1 := registry.Register(service.NewAuthenticationService(nil, repo))
		registry.Register[authentication.UserInfoService](service.NewUserInfoService(repo, repo2))
		registry.Register[authentication.AuthenticationService](srv1.(authentication.AuthenticationService))
		registry.Register[authentication.LoginStatusService](service.NewJWTLoginStatusService())
	}
}

func (b AppMongoDB) Apps() []any {
	return []any{
		InitApp,
	}
}

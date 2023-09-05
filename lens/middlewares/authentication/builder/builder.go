package builder

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/delivery"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/repository"
	"github.com/rhine-tech/scene/lens/middlewares/authentication/service"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/registry"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

//func CreateAndInitialize(cfg model.DatabaseConfig, logger adapter.ILogger) (
//	authentication.AuthenticationService,
//	authentication.LoginStatusService) {
//	repo := globals.Register(repository.NewMongoAuthenticationRepository(cfg))
//	srv1 := globals.Register(service.NewAuthenticationService(logger, repo))
//	globals.Register(srv1.(authentication.AuthenticationService))
//	srv2 := globals.Register(service.NewJWTLoginStatusService())
//	return srv1, srv2
//}

func CreateApp(logger logger.ILogger,
	authSrv authentication.AuthenticationService,
	lgstSrv authentication.LoginStatusService,
	infoSrv authentication.UserInfoService) sgin.GinApplication {
	return delivery.NewGinApp(logger, lgstSrv, authSrv, infoSrv)
}

// Init is instance of scene.LensInit
func Init() {
	cfg := *registry.AcquireSingleton(&model.DatabaseConfig{})
	repo := registry.Register(repository.NewMongoAuthenticationRepository(cfg))
	repo2 := registry.Register(repository.NewUserInfoRepository(cfg))
	srv1 := registry.Register(service.NewAuthenticationService(nil, nil))
	registry.Register(service.NewUserInfoService(repo, repo2))
	registry.Register(srv1.(authentication.AuthenticationService))
	registry.Register(service.NewJWTLoginStatusService())
	return
}

func InitApp() sgin.GinApplication {
	return CreateApp(
		registry.AcquireSingleton(logger.ILogger(nil)),
		registry.AcquireSingleton(authentication.AuthenticationService(nil)),
		registry.AcquireSingleton(authentication.LoginStatusService(nil)),
		registry.AcquireSingleton(authentication.UserInfoService(nil)))
}

type Builder struct {
	scene.Builder
}

func (b Builder) Init() scene.LensInit {
	return Init
}

func (b Builder) Apps() []any {
	return []any{
		InitApp,
	}
}

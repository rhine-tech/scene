package factory

//type GinAppMongoDB struct {
//	scene.ModuleFactory
//	Verifier scene.IModuleDependencyProvider[authentication.HTTPLoginStatusVerifier]
//}
//
//func (b GinAppMongoDB) Default() GinAppMongoDB {
//	return GinAppMongoDB{
//		Verifier: JWTVerifier{
//			Key:    "scene_token",
//			Secret: []byte(registry.Config.GetString("authentication.jwt.secret")),
//		},
//	}
//}
//
//func (b GinAppMongoDB) Init() scene.LensInit {
//	return func() {
//		repo := registry.Load(repository.NewMongoAuthenticationRepository(nil))
//		repo2 := registry.Load(repository.NewUserInfoRepository(nil))
//		srv1 := registry.Register(service.NewAuthenticationService(nil, repo))
//		registry.Register[authentication.UserInfoService](service.NewUserInfoService(repo, repo2))
//		registry.Register[authentication.IAuthenticationService](srv1.(authentication.IAuthenticationService))
//		registry.Register[authentication.HTTPLoginStatusVerifier](b.Verifier.Provide())
//	}
//}
//
//func (b GinAppMongoDB) Apps() []any {
//	return []any{
//		func() sgin.GinApplication {
//			return registry.Load(delivery.NewGinApp(
//				b.Verifier.Provide(),
//				nil,
//				nil))
//		},
//	}
//}

// Package spa declares a static single-page application as a Scene module.
// Its Gin application remains owned by ModuleLoader and selects a named Router;
// it does not register routes outside that Router.
//
// # Configuration
//
// Declare each SPA in the module factory array, using its embedded files:
//
//	builders := scene.ModuleFactoryArray{
//		// Other business modules...
//		spa.SPA{Embed: &mainFiles, FsPrefix: "dist", Router: "web"},
//		spa.SPA{Embed: &adminFiles, FsPrefix: "dist", Router: "admin"},
//	}
//	loader := scene.NewModuleLoader(builders)
//	engine := engines.NewEngine(loader,
//		sgin.NewFactory(":8080",
//			sgin.DefaultRouter("/api", sgin.WithRecovery()),
//			sgin.Router("admin", "/spa", sgin.WithRecovery()),
//			sgin.Router("web", "/", sgin.WithRecovery()),
//		),
//	)
//
// Each SPA should have its own Router because it installs that Router's NoRoute
// handler. GET and HEAD requests for route-style paths fall back to index.html;
// missing static assets return 404. Unknown API paths remain API responses and
// never reach either SPA. Frontend asset URLs must use the deployed URL prefix.
//
// The former SPA.UrlPrefix is replaced by SPA.Router: the URL prefix is now
// configured only in sgin.Router. Likewise, delivery.NewGinSPA's second
// argument is a Router name, not a URL prefix. Its files argument accepts
// fs.FS, including embed.FS and dynamically constructed test files.
package spa

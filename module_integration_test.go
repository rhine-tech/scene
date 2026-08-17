package scene_test

import (
	"path/filepath"
	"testing"

	"github.com/rhine-tech/scene"
	ormfactory "github.com/rhine-tech/scene/composition/orm/factory"
	cachefactory "github.com/rhine-tech/scene/infrastructure/cache/factory"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	datasourcefactory "github.com/rhine-tech/scene/infrastructure/datasource/factory"
	"github.com/rhine-tech/scene/infrastructure/logger"
	loggerfactory "github.com/rhine-tech/scene/infrastructure/logger/factory"
	authfactory "github.com/rhine-tech/scene/lens/authentication/factory"
	permissionfactory "github.com/rhine-tech/scene/lens/permission/factory"
	storagefactory "github.com/rhine-tech/scene/lens/storage/factory"
	sgin "github.com/rhine-tech/scene/scenes/gin"
	"github.com/stretchr/testify/require"
)

func TestModuleLoaderBuildsExistingModules(t *testing.T) {
	loader := scene.NewModuleLoader(scene.ModuleFactoryArray{
		loggerfactory.ZapFactory{LogLevel: logger.LogLevelError, Prefix: "test"},
		datasourcefactory.Sqlite{Config: datasource.SqliteConfig{
			Path:    filepath.Join(t.TempDir(), "scene.db"),
			Options: "mode=rwc",
		}},
		cachefactory.MemoryCache{},
		ormfactory.GormSqlite{},
		authfactory.ServiceGorm{},
		permissionfactory.ServiceGorm{},
		storagefactory.Service{
			Providers:      []storagefactory.StorageProvider{storagefactory.Local{Root: t.TempDir()}},
			SessionTracker: storagefactory.SessionTrackerMemory{},
		},
		authfactory.AppGin{Verifier: authfactory.JWTVerifier{Key: "scene_token", Secret: []byte("test-secret")}},
		permissionfactory.AppGin{},
		storagefactory.App{},
	})
	t.Cleanup(func() { require.NoError(t, loader.TearDown()) })

	require.NoError(t, loader.Init())
	require.NoError(t, loader.Setup())
	applications := scene.Applications[sgin.GinApplication](loader.Applications())
	require.Len(t, applications, 3)
}

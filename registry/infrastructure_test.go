package registry

import (
	"testing"

	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/stretchr/testify/require"
)

type registryConfigProvider struct {
	values map[string]string
}

func (p registryConfigProvider) Init() error {
	return nil
}

func (p registryConfigProvider) GetString(key string) string {
	v, _ := p.GetStringE(key)
	return v
}

func (p registryConfigProvider) GetInt(key string) int64 {
	return 0
}

func (p registryConfigProvider) GetBool(key string) bool {
	return false
}

func (p registryConfigProvider) GetStringE(key string) (string, bool) {
	v, ok := p.values[key]
	return v, ok
}

func (p registryConfigProvider) GetIntE(key string) (int64, bool) {
	return 0, false
}

func (p registryConfigProvider) GetBoolE(key string) (bool, bool) {
	return false, false
}

func TestRegisterConfigRegistersIConfigAndLegacyConfigUnmarshaler(t *testing.T) {
	resetSingletonRegistryForTest()

	cfg := config.NewConfigUnmarshaler(registryConfigProvider{
		values: map[string]string{"scene.name": "Scene"},
	})
	RegisterConfig(cfg)

	require.Same(t, cfg, Config)
	require.Equal(t, "Scene", Config.GetString("scene.name"))
	require.Equal(t, "fallback", Config.GetStringOrDefault("missing", "fallback"))

	iConfig := AcquireSingleton(config.IConfig(nil))
	require.Same(t, cfg, iConfig)

	legacy := AcquireSingleton(config.ConfigUnmarshaler(nil))
	require.Same(t, cfg, legacy)

	withDefault := AcquireSingleton(config.ConfigProviderWithDefault(nil))
	require.Same(t, cfg, withDefault)
	require.Equal(t, "fallback", withDefault.GetStringOrDefault("missing", "fallback"))
}

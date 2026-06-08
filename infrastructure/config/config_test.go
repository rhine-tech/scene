package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mapConfigProvider struct {
	strings map[string]string
	ints    map[string]int64
	bools   map[string]bool
}

func (p mapConfigProvider) Init() error {
	return nil
}

func (p mapConfigProvider) GetString(key string) string {
	v, _ := p.GetStringE(key)
	return v
}

func (p mapConfigProvider) GetInt(key string) int64 {
	v, _ := p.GetIntE(key)
	return v
}

func (p mapConfigProvider) GetBool(key string) bool {
	v, _ := p.GetBoolE(key)
	return v
}

func (p mapConfigProvider) GetStringE(key string) (string, bool) {
	v, ok := p.strings[key]
	return v, ok
}

func (p mapConfigProvider) GetIntE(key string) (int64, bool) {
	v, ok := p.ints[key]
	return v, ok
}

func (p mapConfigProvider) GetBoolE(key string) (bool, bool) {
	v, ok := p.bools[key]
	return v, ok
}

func TestUnmarshalDoesNotPrefixEmptyKey(t *testing.T) {
	type AppConfig struct {
		Host string `scfg:"host"`
	}

	cfg := NewConfigUnmarshaler(mapConfigProvider{
		strings: map[string]string{"host": "localhost"},
	})
	var appConfig AppConfig
	require.NoError(t, cfg.Unmarshal(&appConfig))
	require.Equal(t, "localhost", appConfig.Host)
}

func TestUnmarshalWithPrefixReadsPrefixedKey(t *testing.T) {
	type SqliteConfig struct {
		Path string `scfg:"path"`
	}

	cfg := NewConfigUnmarshaler(mapConfigProvider{
		strings: map[string]string{"sqlite.path": "data.db"},
	})
	var sqliteConfig SqliteConfig
	require.NoError(t, cfg.UnmarshalWithPrefix("sqlite", &sqliteConfig))
	require.Equal(t, "data.db", sqliteConfig.Path)
}

func TestUnmarshalWithOptionsUsesScfgTagByDefault(t *testing.T) {
	type AppConfig struct {
		Timeout int64 `scfg:"timeout,default=30"`
	}

	var cfg AppConfig
	require.NoError(t, UnmarshalWithOptions(mapConfigProvider{}, &cfg))
	require.Equal(t, int64(30), cfg.Timeout)
}

func TestUnmarshalWithOptionsNilProviderReturnsError(t *testing.T) {
	type AppConfig struct {
		Timeout int64 `scfg:"timeout,default=30"`
	}

	err := UnmarshalWithOptions(nil, &AppConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "provider is nil")
}

func TestUnmarshalWithOptionsDefaultBehaviorZeroValue(t *testing.T) {
	type AppConfig struct {
		Timeout int64 `scfg:"timeout,default=30"`
	}

	var cfg AppConfig
	require.NoError(t, UnmarshalWithOptions(
		mapConfigProvider{},
		&cfg,
		WithDefaultBehavior(DefaultBehaviorZeroValue),
	))
	require.Equal(t, int64(0), cfg.Timeout)
}

func TestUnmarshalWithOptionsIntKindsAndPrefix(t *testing.T) {
	type AppConfig struct {
		Value   int   `scfg:"value"`
		Small   int8  `scfg:"small"`
		Medium  int32 `scfg:"medium"`
		Timeout int64 `scfg:"timeout,default=30"`
	}

	var cfg AppConfig
	provider := mapConfigProvider{
		ints: map[string]int64{
			"app.value":  10,
			"app.small":  8,
			"app.medium": 32,
		},
	}

	require.NoError(t, UnmarshalWithOptions(provider, &cfg, WithPrefix("app")))
	require.Equal(t, 10, cfg.Value)
	require.Equal(t, int8(8), cfg.Small)
	require.Equal(t, int32(32), cfg.Medium)
	require.Equal(t, int64(30), cfg.Timeout)
}

func TestUnmarshalWithOptionsUnsupportedTypeReturnsError(t *testing.T) {
	type AppConfig struct {
		Values []string `scfg:"values"`
	}

	err := UnmarshalWithOptions(mapConfigProvider{}, &AppConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported config field type")
}

func TestUnmarshalWithOptionsIntOverflowReturnsError(t *testing.T) {
	type AppConfig struct {
		Small int8 `scfg:"small"`
	}

	err := UnmarshalWithOptions(mapConfigProvider{
		ints: map[string]int64{"small": 128},
	}, &AppConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "overflows")
}

func TestDefaultValueAPI(t *testing.T) {
	provider := mapConfigProvider{
		strings: map[string]string{"name": "scene"},
		ints:    map[string]int64{"port": 8080},
		bools:   map[string]bool{"debug": true},
	}

	require.Equal(t, "scene", GetStringOrDefault(provider, "name", "fallback"))
	require.Equal(t, "fallback", GetStringOrDefault(provider, "missing", "fallback"))
	require.Equal(t, "fallback", GetStringOrDefault(nil, "missing", "fallback"))

	require.EqualValues(t, 8080, GetIntOrDefault(provider, "port", 80))
	require.EqualValues(t, 80, GetIntOrDefault(provider, "missing", 80))
	require.EqualValues(t, 80, GetIntOrDefault(nil, "missing", 80))

	require.True(t, GetBoolOrDefault(provider, "debug", false))
	require.True(t, GetBoolOrDefault(provider, "missing", true))
	require.True(t, GetBoolOrDefault(nil, "missing", true))
}

func TestNewConfigUnmarshalerReturnsIConfig(t *testing.T) {
	cfg := NewConfigUnmarshaler(mapConfigProvider{
		strings: map[string]string{"name": "scene"},
	})

	var iConfig IConfig = cfg
	require.Equal(t, "scene", iConfig.GetStringOrDefault("name", "fallback"))
	require.Equal(t, "fallback", iConfig.GetStringOrDefault("missing", "fallback"))
}

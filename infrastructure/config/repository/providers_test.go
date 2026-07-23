package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/stretchr/testify/require"
)

func testProvidersWithSameKeys(t *testing.T) map[string]config.IConfig {
	t.Helper()

	dir := t.TempDir()

	tomlFile := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(tomlFile, []byte(`[server]
host = "localhost"
port = 8080
debug = true
bad_int = "not-an-int"
bad_bool = "not-a-bool"
`), 0644))

	jsonFile := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(jsonFile, []byte(`{
  "server": {
    "host": "localhost",
    "port": 8080,
    "debug": true,
    "bad_int": "not-an-int",
    "bad_bool": "not-a-bool"
  }
}`), 0644))

	iniFile := filepath.Join(dir, "config.ini")
	require.NoError(t, os.WriteFile(iniFile, []byte(`[server]
host = localhost
port = 8080
debug = true
bad_int = not-an-int
bad_bool = not-a-bool
`), 0644))

	dotenvFile := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(dotenvFile, []byte(`server_host=localhost
server_port=8080
server_debug=true
server_bad_int=not-an-int
server_bad_bool=not-a-bool
`), 0644))

	t.Setenv("server_host", "localhost")
	t.Setenv("server_port", "8080")
	t.Setenv("server_debug", "true")
	t.Setenv("server_bad_int", "not-an-int")
	t.Setenv("server_bad_bool", "not-a-bool")

	return map[string]config.IConfig{
		"toml":   NewTomlConfig(tomlFile),
		"json":   NewJsonMarshaller(jsonFile),
		"ini":    NewIniConfig(iniFile),
		"dotenv": NewDotenvMarshaller(dotenvFile),
		"env":    NewEnvMarshaller(),
	}
}

func TestProvidersReadSameKeysConsistently(t *testing.T) {
	for name, cfg := range testProvidersWithSameKeys(t) {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, cfg.Init())

			require.Equal(t, "localhost", cfg.GetString("server.host"))
			require.EqualValues(t, 8080, cfg.GetInt("server.port"))
			require.True(t, cfg.GetBool("server.debug"))

			val, ok := cfg.GetStringE("server.host")
			require.True(t, ok)
			require.Equal(t, "localhost", val)

			intVal, ok := cfg.GetIntE("server.port")
			require.True(t, ok)
			require.EqualValues(t, 8080, intVal)

			boolVal, ok := cfg.GetBoolE("server.debug")
			require.True(t, ok)
			require.True(t, boolVal)

			_, ok = cfg.GetStringE("server.missing")
			require.False(t, ok)
			_, ok = cfg.GetIntE("server.missing")
			require.False(t, ok)
			_, ok = cfg.GetBoolE("server.missing")
			require.False(t, ok)

			_, ok = cfg.GetIntE("server.bad_int")
			require.False(t, ok)
			_, ok = cfg.GetBoolE("server.bad_bool")
			require.False(t, ok)

			require.Equal(t, "fallback", config.GetStringOrDefault(cfg, "server.missing", "fallback"))
			require.EqualValues(t, 80, config.GetIntOrDefault(cfg, "server.missing", 80))
			require.True(t, config.GetBoolOrDefault(cfg, "server.missing", true))
		})
	}
}

func TestTomlConfigGettersAndUnmarshal(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(filename, []byte(`[server]
host = "localhost"
port = 8080
debug = true

[database]
user = "root"
port = 3306
`), 0644))

	cfg := NewTomlConfig(filename)
	require.NoError(t, cfg.Init())

	require.Equal(t, "localhost", cfg.GetString("server.host"))
	require.EqualValues(t, 8080, cfg.GetInt("server.port"))
	require.True(t, cfg.GetBool("server.debug"))

	type AppConfig struct {
		Host       string `scfg:"server.host,default=127.0.0.1"`
		Port       int    `scfg:"server.port,default=80"`
		DBUser     string `scfg:"database.user,default=admin"`
		MissingStr string `scfg:"feature.missing,default=hello"`
	}
	var appConfig AppConfig
	require.NoError(t, cfg.Unmarshal(&appConfig))
	require.Equal(t, "localhost", appConfig.Host)
	require.Equal(t, 8080, appConfig.Port)
	require.Equal(t, "root", appConfig.DBUser)
	require.Equal(t, "hello", appConfig.MissingStr)
}

func TestTomlConfigInitFileNotExist(t *testing.T) {
	cfg := NewTomlConfig(filepath.Join(t.TempDir(), "not_exist.toml"))
	require.Error(t, cfg.Init())
}

func TestNewTomlCfgurUnmarshalWithScfgTags(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(filename, []byte(`[scene]
name = "xiyin"
`), 0644))

	cfg := NewTomlCfgur(filename)
	require.NoError(t, cfg.Init())

	type SceneConfig struct {
		Name string `scfg:"name"`
	}
	var sceneConfig SceneConfig
	require.NoError(t, cfg.UnmarshalWithPrefix("scene", &sceneConfig))
	require.Equal(t, "xiyin", sceneConfig.Name)
}

func TestJsonConfigGettersAndUnmarshal(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(filename, []byte(`{
  "server": {
    "host": "localhost",
    "port": 8080,
    "debug": true
  }
}`), 0644))

	cfg := NewJsonMarshaller(filename)
	require.NoError(t, cfg.Init())

	require.Equal(t, "localhost", cfg.GetString("server.host"))
	require.EqualValues(t, 8080, cfg.GetInt("server.port"))
	require.True(t, cfg.GetBool("server.debug"))

	type AppConfig struct {
		Host    string `scfg:"server.host"`
		Missing string `scfg:"server.missing,default=fallback"`
	}
	var appConfig AppConfig
	require.NoError(t, cfg.Unmarshal(&appConfig))
	require.Equal(t, "localhost", appConfig.Host)
	require.Equal(t, "fallback", appConfig.Missing)
}

func TestJsonConfigInvalidIntAndBoolAreNotValidValues(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(filename, []byte(`{
  "server": {
    "port": "not-an-int",
    "debug": "not-a-bool"
  }
}`), 0644))

	cfg := NewJsonMarshaller(filename)
	require.NoError(t, cfg.Init())

	_, ok := cfg.GetIntE("server.port")
	require.False(t, ok)
	_, ok = cfg.GetBoolE("server.debug")
	require.False(t, ok)
}

func TestIniConfigMissingStringUsesDefault(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.ini")
	require.NoError(t, os.WriteFile(filename, []byte(`[server]
http_port = 9999
`), 0644))

	cfg := NewIniConfig(filename)
	require.NoError(t, cfg.Init())

	type AppConfig struct {
		Host string `scfg:"server.host,default=127.0.0.1"`
	}
	var appConfig AppConfig
	require.NoError(t, cfg.Unmarshal(&appConfig))
	require.Equal(t, "127.0.0.1", appConfig.Host)
}

func TestIniMissingIntAndBoolReadsDoNotCreateStringKey(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.ini")
	require.NoError(t, os.WriteFile(filename, []byte(`[server]
host = localhost
`), 0644))

	cfg := NewIniConfig(filename)
	require.NoError(t, cfg.Init())

	_, ok := cfg.GetIntE("server.missing")
	require.False(t, ok)
	_, ok = cfg.GetBoolE("server.missing")
	require.False(t, ok)
	require.Equal(t, "fallback", config.GetStringOrDefault(cfg, "server.missing", "fallback"))
}

func TestEnvMarshallerUnmarshal(t *testing.T) {
	t.Setenv("username", "test")
	t.Setenv("value", "1")
	t.Setenv("enabled", "true")

	type AppConfig struct {
		Username string `scfg:"username"`
		Value    int    `scfg:"value"`
		Enabled  bool   `scfg:"enabled"`
		Val2     int    `scfg:"val2,default=3"`
	}

	var appConfig AppConfig
	cfg := NewEnvMarshaller()
	require.NoError(t, cfg.Unmarshal(&appConfig))
	require.Equal(t, "test", appConfig.Username)
	require.Equal(t, 1, appConfig.Value)
	require.True(t, appConfig.Enabled)
	require.Equal(t, 3, appConfig.Val2)
}

func TestEnvMarshallerInvalidIntAndBoolAreNotValidValues(t *testing.T) {
	t.Setenv("bad_int", "not-an-int")
	t.Setenv("bad_bool", "not-a-bool")

	cfg := NewEnvMarshaller()
	_, ok := cfg.GetIntE("bad.int")
	require.False(t, ok)
	_, ok = cfg.GetBoolE("bad.bool")
	require.False(t, ok)
}

func TestCompatibilityConstructors(t *testing.T) {
	tomlFile := filepath.Join(t.TempDir(), "config.toml")
	require.NoError(t, os.WriteFile(tomlFile, []byte(`[scene]
name = "xiyin"
`), 0644))

	jsonFile := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(jsonFile, []byte(`{"scene":{"name":"xiyin"}}`), 0644))

	iniFile := filepath.Join(t.TempDir(), "config.ini")
	require.NoError(t, os.WriteFile(iniFile, []byte(`[scene]
name = xiyin
`), 0644))

	dotenvFile := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(dotenvFile, []byte("scene_name=xiyin\n"), 0644))

	constructors := []struct {
		name string
		cfg  interface {
			Init() error
			GetString(string) string
		}
	}{
		{name: "toml", cfg: NewTomlCfgur(tomlFile)},
		{name: "json", cfg: NewJsonCfgur(jsonFile)},
		{name: "ini", cfg: NewINICfgur(iniFile)},
		{name: "dotenv", cfg: NewDotEnvironmentCfgur(dotenvFile)},
	}

	for _, constructor := range constructors {
		t.Run(constructor.name, func(t *testing.T) {
			require.NoError(t, constructor.cfg.Init())
			require.Equal(t, "xiyin", constructor.cfg.GetString("scene.name"))
		})
	}

	t.Setenv("scene_name", "xiyin")
	cfg := NewEnvironmentCfgur()
	require.NoError(t, cfg.Init())
	require.Equal(t, "xiyin", cfg.GetString("scene.name"))
}

func TestDotenvMarshallerInvalidIntAndBoolAreNotValidValues(t *testing.T) {
	filename := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(filename, []byte("bad_int=not-an-int\nbad_bool=not-a-bool\n"), 0644))

	cfg := NewDotenvMarshaller(filename)
	require.NoError(t, cfg.Init())

	_, ok := cfg.GetIntE("bad.int")
	require.False(t, ok)
	_, ok = cfg.GetBoolE("bad.bool")
	require.False(t, ok)
}

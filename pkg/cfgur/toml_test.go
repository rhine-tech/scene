package cfgur

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestTomlConfig(t *testing.T) ConfigUnmarshaler {
	t.Helper()

	cfg := NewTomlConfig("testdata/config.toml")
	// NewTomlConfig 返回的是 *commonMarshaller
	require.NoError(t, cfg.Init())
	return cfg
}

func TestTomlConfig_Getters(t *testing.T) {
	cfg := newTestTomlConfig(t)

	// 基础类型读取
	require.Equal(t, "localhost", cfg.GetString("server.host"))
	require.EqualValues(t, 8080, cfg.GetInt("server.port"))
	require.Equal(t, true, cfg.GetBool("server.debug"))

	// GetStringE / GetIntE / GetBoolE 正常情况
	s, ok := cfg.GetStringE("database.user")
	require.True(t, ok)
	require.Equal(t, "root", s)

	port, ok := cfg.GetIntE("database.port")
	require.True(t, ok)
	require.EqualValues(t, 3306, port)

	flag, ok := cfg.GetBoolE("feature.flag")
	require.True(t, ok)
	require.Equal(t, false, flag)

	// 不存在的 key 返回 false
	_, ok = cfg.GetStringE("no.such.key")
	require.False(t, ok)

	_, ok = cfg.GetIntE("server.no_port")
	require.False(t, ok)

	_, ok = cfg.GetBoolE("feature.not_exist_flag")
	require.False(t, ok)
}

func TestTomlConfig_Unmarshal_WithTagsAndDefaults(t *testing.T) {
	cfg := newTestTomlConfig(t)

	type AppConfig struct {
		Host        string `cfgur:"server.host,default=127.0.0.1"`
		Port        int    `cfgur:"server.port,default=80"`
		Debug       bool   `cfgur:"server.debug,default=false"`
		DBUser      string `cfgur:"database.user,default=admin"`
		MissingStr  string `cfgur:"feature.missing,default=hello"`
		MissingBool bool   `cfgur:"feature.missing_bool,default=true"`
		MissingInt  int    `cfgur:"feature.missing_int,default=42"`
	}

	var ac AppConfig
	require.NoError(t, cfg.Unmarshal(&ac))

	// 有值的字段应该来自 config.toml
	require.Equal(t, "localhost", ac.Host)
	require.Equal(t, 8080, ac.Port)
	require.Equal(t, true, ac.Debug)
	require.Equal(t, "root", ac.DBUser)

	// 没有在 config.toml 中出现的字段使用默认值
	require.Equal(t, "hello", ac.MissingStr)
	require.Equal(t, true, ac.MissingBool)
	require.Equal(t, 42, ac.MissingInt)
}

func TestTomlConfig_Init_FileNotExist(t *testing.T) {
	cfg := NewTomlConfig("testdata/not_exist.toml")
	err := cfg.Init()
	require.Error(t, err, "Init 应该在文件不存在时返回错误")
}

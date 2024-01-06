package cfgur

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewIniConfig(t *testing.T) {
	cfg := NewIniConfig("ini_config.ini")
	require.NoError(t, cfg.Init())
	require.Equal(t, "development", cfg.GetString("app_mode"))
	require.Equal(t, int64(9999), cfg.GetInt("server.http_port"))
	require.Equal(t, true, cfg.GetBool("server.enforce_domain"))
}

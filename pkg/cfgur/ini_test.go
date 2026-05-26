package cfgur

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestNewIniConfig(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "ini_config.ini")
	require.NoError(t, os.WriteFile(filename, []byte(`app_mode = development

[server]
http_port = 9999
enforce_domain = true
`), 0644))
	cfg := NewIniConfig(filename)
	require.NoError(t, cfg.Init())
	require.Equal(t, "development", cfg.GetString("app_mode"))
	require.Equal(t, int64(9999), cfg.GetInt("server.http_port"))
	require.Equal(t, true, cfg.GetBool("server.enforce_domain"))
}

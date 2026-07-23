package datasource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMysqlConfigMaskedDSN(t *testing.T) {
	cfg := MysqlConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "scene",
		Password: "secret",
		Database: "scene",
		Options:  "parseTime=true",
	}

	require.Equal(
		t,
		"scene:REDACTED@tcp(localhost:3306)/scene?parseTime=true",
		cfg.MaskedDSN(),
	)
	require.NotContains(t, cfg.MaskedDSN(), cfg.Password)
}

func TestRedisConfigMaskedDSN(t *testing.T) {
	cfg := RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Username: "scene",
		Password: "secret",
		Database: 2,
	}

	require.Equal(
		t,
		"redis://scene:REDACTED@localhost:6379/2",
		cfg.MaskedDSN(),
	)
	require.NotContains(t, cfg.MaskedDSN(), cfg.Password)
}

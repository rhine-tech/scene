package datasource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMongoConfigDSNWithAuthSource(t *testing.T) {
	cfg := MongoConfig{
		Host:       "localhost",
		Port:       27017,
		Username:   "scene",
		Password:   "secret",
		AuthSource: "admin db",
		Options:    "replicaSet=rs0",
	}

	require.Equal(
		t,
		"mongodb://scene:secret@localhost:27017/?replicaSet=rs0&authSource=admin+db",
		cfg.DSN(),
	)
}

func TestMongoConfigDSNEncodesCredentials(t *testing.T) {
	cfg := MongoConfig{
		Host:     "localhost",
		Port:     27017,
		Username: "scene@example.com",
		Password: "p@ss:/word",
	}

	require.Equal(
		t,
		"mongodb://scene%40example.com:p%40ss%3A%2Fword@localhost:27017/",
		cfg.DSN(),
	)
}

func TestMongoConfigMaskedDSN(t *testing.T) {
	cfg := MongoConfig{
		Host:     "localhost",
		Port:     27017,
		Username: "scene",
		Password: "secret",
	}

	require.Equal(
		t,
		"mongodb://scene:REDACTED@localhost:27017/",
		cfg.MaskedDSN(),
	)
	require.NotContains(t, cfg.MaskedDSN(), cfg.Password)
}

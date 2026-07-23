package datasource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresConfigDSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  PostgresConfig
		want string
	}{
		{
			name: "credentials and options",
			cfg: PostgresConfig{
				Host:     "postgres.example.com",
				Port:     5432,
				Username: "scene",
				Password: "secret",
				Database: "scene",
				Options:  "sslmode=require&application_name=scene",
			},
			want: "postgres://scene:secret@postgres.example.com:5432/scene?sslmode=require&application_name=scene",
		},
		{
			name: "escaped credentials and ipv6",
			cfg: PostgresConfig{
				Host:     "2001:db8::1",
				Port:     5432,
				Username: "scene@example.com",
				Password: "p@ss:word",
				Database: "scene data",
			},
			want: "postgres://scene%40example.com:p%40ss%3Aword@[2001:db8::1]:5432/scene%20data",
		},
		{
			name: "without credentials",
			cfg: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "scene",
			},
			want: "postgres://localhost:5432/scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.cfg.DSN())
		})
	}
}

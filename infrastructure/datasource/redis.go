package datasource

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type RedisConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database int
}

func (c RedisConfig) DSN() string {
	var sb strings.Builder
	sb.WriteString("redis://")
	if c.Username != "" {
		sb.WriteString(fmt.Sprintf("%s:%s@", c.Username, c.Password))
	}
	sb.WriteString(fmt.Sprintf("%s:%d/%d", c.Host, c.Port, c.Database))
	return sb.String()
}

func (c RedisConfig) MaskedDSN() string {
	if c.Password != "" {
		c.Password = maskedPassword
	}
	return c.DSN()
}

type RedisDataSource interface {
	DataSource
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys ...string) ([]string, error)
	Incr(ctx context.Context, key string) (int64, error)
	GetValue(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
}

package repository

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/model"
	"github.com/spf13/cast"
	"time"
)

type RedisDataRepo struct {
	rdb *redis.Client
}

func NewRedisDataRepo(cfg model.DatabaseConfig) datasource.RedisDataSource {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		Username: cfg.Username,
		DB:       cast.ToInt(cfg.Database),
	})
	return &RedisDataRepo{rdb: rdb}
}

func (r *RedisDataRepo) DataSourceName() string {
	return "datasource.repository.redis"
}

func (r *RedisDataRepo) Setup() error {
	if err := r.Status(); err != nil {
		return err
	}
	return nil
}

func (r *RedisDataRepo) Status() error {
	return r.rdb.Ping(context.Background()).Err()
}

func (r *RedisDataRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}

func (r *RedisDataRepo) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(context.Background(), key).Result()
}

func (r *RedisDataRepo) Delete(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisDataRepo) Dispose() error {
	return r.rdb.Close()
}

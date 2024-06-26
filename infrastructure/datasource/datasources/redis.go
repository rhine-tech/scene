package datasources

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/model"
	"github.com/spf13/cast"
	"time"
)

type RedisDataRepo struct {
	cfg model.DatabaseConfig
	rdb *redis.Client
	log logger.ILogger `aperture:""`
}

func NewRedisDataRepo(cfg model.DatabaseConfig) datasource.RedisDataSource {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		Username: cfg.Username,
		DB:       cast.ToInt(cfg.Database),
	})
	return &RedisDataRepo{rdb: rdb, cfg: cfg}
}

func (r *RedisDataRepo) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("RedisDataSource")
}

func (r *RedisDataRepo) Setup() error {
	r.log = r.log.WithPrefix(r.DataSourceName().String())
	if err := r.Status(); err != nil {
		r.log.Errorf("establish connection '%s' failed", r.cfg.RedisDSN())
		return err
	}
	r.log.Infof("establish connection '%s' succeed", r.cfg.RedisDSN())
	return nil
}

func (r *RedisDataRepo) Dispose() error {
	err := r.rdb.Close()
	if err != nil {
		r.log.Warnf("close '%s' failed", r.cfg.RedisDSN())
	}
	r.log.Infof("close connection '%s' succeed", r.cfg.RedisDSN())
	return err
}

func (r *RedisDataRepo) Status() error {
	return r.rdb.Ping(context.Background()).Err()
}

func (r *RedisDataRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}

func (r *RedisDataRepo) GetValue(ctx context.Context, key string, value interface{}) error {
	_, isJsonUnmarshallable := value.(json.Unmarshaler)
	_, isBinaryUnmarshallable := value.(encoding.BinaryUnmarshaler)
	if isJsonUnmarshallable && !isBinaryUnmarshallable {
		return r.rdb.Get(ctx, key).Scan(&binary2jsonMarshaller{value})
	}
	return r.rdb.Get(ctx, key).Scan(value)
}

func (r *RedisDataRepo) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(context.Background(), key).Result()
}

func (r *RedisDataRepo) Delete(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

type binary2jsonMarshaller struct {
	value interface{}
}

func (i *binary2jsonMarshaller) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i.value)
}

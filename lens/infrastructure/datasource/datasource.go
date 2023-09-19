package datasource

import (
	"context"
	"database/sql"
	"github.com/rhine-tech/scene"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type DataSource interface {
	scene.Disposable
	scene.Setupable
	DataSourceName() string
	Status() error
}

type MongoDataSource interface {
	DataSource
	Database() *mongo.Database
	Collection(coll string) *mongo.Collection
}

type MysqlDataSource interface {
	DataSource
	Connection() *sql.DB
}

type RedisDataSource interface {
	DataSource
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

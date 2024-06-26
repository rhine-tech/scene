package datasource

import (
	"context"
	"database/sql"
	"github.com/rhine-tech/scene"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const Lens scene.InfraName = "datasource"

type DataSource interface {
	scene.Disposable
	scene.Setupable
	DataSourceName() scene.ImplName
	Status() error
}

type MongoDataSource interface {
	DataSource
	Database() *mongo.Database
	Collection(coll string) *mongo.Collection
}

type SqlDataSource interface {
	DataSource
	Connection() *sql.DB
}

type MysqlDataSource interface {
	SqlDataSource
}

type SqliteDataSource interface {
	SqlDataSource
}

type RedisDataSource interface {
	DataSource
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetValue(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
}

type JsonDataSource interface {
	DataSource
	Load() ([]byte, error)
	Save(data []byte) error
}

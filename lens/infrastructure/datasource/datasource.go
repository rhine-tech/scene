package datasource

import (
	"database/sql"
	"github.com/aynakeya/scene"
	"go.mongodb.org/mongo-driver/mongo"
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

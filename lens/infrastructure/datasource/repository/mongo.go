package repository

import (
	"context"
	"fmt"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoRepo struct {
	Cfg    model.DatabaseConfig
	client *mongo.Client
	err    error
	db     *mongo.Database
}

var _ datasource.MongoDataSource = (*MongoRepo)(nil)

func NewMongoDataSource(cfg model.DatabaseConfig) datasource.MongoDataSource {
	repo := &MongoRepo{Cfg: cfg}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	var uri string
	if cfg.Username == "" {
		uri = fmt.Sprintf("mongodb://%s:%d/", cfg.Host, cfg.Port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	}
	opts := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	repo.client, repo.err = mongo.Connect(context.TODO(), opts)
	if repo.err != nil {
		return repo
	}
	repo.db = repo.client.Database(cfg.Database)
	return repo
}
func (m *MongoRepo) DataSourceName() string {
	return "datasource.repository.mongo"
}

func (m *MongoRepo) Status() error {
	if m.err != nil {
		return m.err
	}
	return m.client.Ping(context.Background(), readpref.Primary())
}

func (m *MongoRepo) Database() *mongo.Database {
	if m.db == nil {
		m.db = m.client.Database(m.Cfg.Database)
	}
	return m.db
}

func (m *MongoRepo) Collection(coll string) *mongo.Collection {
	return m.Database().Collection(coll)
}

func (m *MongoRepo) Setup() error {
	return nil
}

func (m *MongoRepo) Dispose() error {
	if m.client != nil {
		return m.client.Disconnect(context.Background())
	}
	m.client = nil
	return nil
}

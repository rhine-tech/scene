package repository

import (
	"context"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoRepo struct {
	cfg    model.DatabaseConfig
	client *mongo.Client
	err    error
	db     *mongo.Database
	log    logger.ILogger `aperture:""`
}

var _ datasource.MongoDataSource = (*MongoRepo)(nil)

func NewMongoDataSource(cfg model.DatabaseConfig) datasource.MongoDataSource {
	repo := &MongoRepo{cfg: cfg}
	return repo
}

func (m *MongoRepo) DataSourceName() string {
	return "datasource.repository.mongo"
}

func (m *MongoRepo) Setup() error {
	m.log = m.log.WithPrefix(m.DataSourceName())
	if m.err != nil {
		m.log.Errorf("%s init failed", m.cfg.MongoDSN())
		return m.err
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI(m.cfg.MongoDSN()).
		SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	m.client, m.err = mongo.Connect(context.TODO(), opts)
	if m.err != nil {
		m.log.Warnf("%s init failed", m.cfg.MongoDSN())
		return m.err
	}
	m.db = m.client.Database(m.cfg.Database)
	m.log.Infof("establish connection to %s succeed", m.cfg.MongoDSN())
	return nil
}

func (m *MongoRepo) Dispose() error {
	if m.client == nil {
		return nil
	}
	err := m.client.Disconnect(context.Background())
	if err != nil {
		m.log.Warnf("%s close failed", m.cfg.MongoDSN())
		return err
	}
	m.log.Infof("close connection %s success", m.cfg.MongoDSN())
	m.client = nil
	return err
}

func (m *MongoRepo) Status() error {
	if m.err != nil {
		return m.err
	}
	return m.client.Ping(context.Background(), readpref.Primary())
}

func (m *MongoRepo) Database() *mongo.Database {
	if m.db == nil {
		m.db = m.client.Database(m.cfg.Database)
	}
	return m.db
}

func (m *MongoRepo) Collection(coll string) *mongo.Collection {
	return m.Database().Collection(coll)
}

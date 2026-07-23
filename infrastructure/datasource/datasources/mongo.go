package datasources

import (
	"context"
	"errors"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var errMongoDataSourceNotSetup = errors.New("mongo datasource is not set up")

type MongoRepo struct {
	cfg    datasource.MongoConfig
	client *mongo.Client
	err    error
	db     *mongo.Database
	log    logger.ILogger `aperture:""`
}

var _ datasource.MongoDataSource = (*MongoRepo)(nil)

func NewMongoDataSource(cfg datasource.MongoConfig) datasource.MongoDataSource {
	repo := &MongoRepo{cfg: cfg}
	return repo
}

func (j *MongoRepo) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("MongoDataSource")
}

func (m *MongoRepo) Setup() error {
	m.log = m.log.WithPrefix(m.DataSourceName().String())
	if m.err != nil {
		m.log.Errorf("%s init failed", m.cfg.MaskedDSN())
		return m.err
	}
	var opts *options.ClientOptions
	if m.cfg.UseAPIVersion {
		m.log.Infof("use default server api version 1")
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts = options.Client().
			ApplyURI(m.cfg.DSN()).
			SetServerAPIOptions(serverAPI)
	} else {
		m.log.Infof("dont use default server api version for backward compatibility")
		opts = options.Client().ApplyURI(m.cfg.DSN())
	}

	// Create a new client and connect to the server
	m.client, m.err = mongo.Connect(opts)
	if m.err != nil {
		m.log.Warnf("'%s' init failed", m.cfg.MaskedDSN())
		return m.err
	}
	m.db = m.client.Database(m.cfg.Database)
	if m.err = m.Status(); m.err != nil {
		_ = m.client.Disconnect(context.Background())
		m.client = nil
		m.db = nil
		m.log.Warnf("'%s' failed to connect: %s", m.cfg.MaskedDSN(), m.err)
		return m.err
	}
	m.log.Infof("establish connection to '%s' succeed", m.cfg.MaskedDSN())
	return nil
}

func (m *MongoRepo) Dispose() error {
	if m.client == nil {
		return nil
	}
	err := m.client.Disconnect(context.Background())
	if err != nil {
		m.log.Warnf("%s close failed", m.cfg.MaskedDSN())
		return err
	}
	m.log.Infof("close connection %s success", m.cfg.MaskedDSN())
	m.client = nil
	return err
}

func (m *MongoRepo) Status() error {
	if m.err != nil {
		return m.err
	}
	if m.client == nil {
		return errMongoDataSourceNotSetup
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

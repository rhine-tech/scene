package repos

import (
	"context"
	"fmt"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Deprecated: use MongoDatasourceCollection instead
type MongoRepo struct {
	Cfg    model.DatabaseConfig
	client *mongo.Client
	err    error
	db     *mongo.Database
}

func NewMongoRepo(cfg model.DatabaseConfig) *MongoRepo {
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

func (m *MongoRepo) Dispose() error {
	if m.client != nil {
		return m.client.Disconnect(context.Background())
	}
	m.client = nil
	return nil
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

// Deprecated: use MongoDatasourceCollection instead
type MongoCollectionRepo[T any] struct {
	Cfg    model.DatabaseConfig
	client *mongo.Client
	err    error
	db     *mongo.Database
	coll   *mongo.Collection
}

func NewMongoCollectionRepo[T any](cfg model.DatabaseConfig, collection string) *MongoCollectionRepo[T] {
	repo := &MongoCollectionRepo[T]{Cfg: cfg}
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
	if repo.Status() != nil {
		return repo
	}
	repo.coll = repo.db.Collection(collection)
	return repo
}

func (m *MongoCollectionRepo[T]) FindOne(filter interface{}) (T, error) {
	var result T
	err := m.coll.FindOne(context.Background(), filter).Decode(&result)
	return result, err
}

func (m *MongoCollectionRepo[T]) FindPagination(filter interface{}, sort interface{}, offset int64, limit int64) (result []T, total int, err error) {
	opts := options.Find().SetLimit(limit).SetSkip(offset).SetSort(sort)
	cursor, err := m.coll.Find(context.Background(), filter, opts)
	cnt, _ := m.coll.CountDocuments(context.Background(), filter)
	defer cursor.Close(context.Background())
	if err != nil {
		return []T{}, 0, err
	}
	var results []T
	if err = cursor.All(context.Background(), &results); err != nil {
		return []T{}, 0, err
	}
	return results, int(cnt), nil
}

func (m *MongoCollectionRepo[T]) Collection() *mongo.Collection {
	return m.coll
}

func (m *MongoCollectionRepo[T]) Dispose() error {
	if m.client != nil {
		return m.client.Disconnect(context.Background())
	}
	m.client = nil
	return nil
}

func (m *MongoCollectionRepo[T]) Status() error {
	if m.err != nil {
		return m.err
	}
	return m.client.Ping(context.Background(), readpref.Primary())
}

func (m *MongoCollectionRepo[T]) Database() *mongo.Database {
	if m.db == nil {
		m.db = m.client.Database(m.Cfg.Database)
	}
	return m.db
}

type MongoDatasourceCollection[T any] struct {
	datasource     datasource.MongoDataSource
	coll           *mongo.Collection
	CollectionName string
}

func UseMongoDatasourceCollection[T any](datasource datasource.MongoDataSource, coll string) *MongoDatasourceCollection[T] {
	rp := &MongoDatasourceCollection[T]{datasource: datasource}
	rp.coll = datasource.Collection(coll)
	rp.CollectionName = coll
	return rp
}

func (m *MongoDatasourceCollection[T]) Status() error {
	return m.datasource.Status()
}

func (m *MongoDatasourceCollection[T]) Collection() *mongo.Collection {
	return m.coll
}

func (m *MongoDatasourceCollection[T]) FindOne(filter interface{}) (T, error) {
	var result T
	err := m.coll.FindOne(context.Background(), filter).Decode(&result)
	return result, err
}

func (m *MongoDatasourceCollection[T]) FindPagination(filter interface{}, sort interface{}, offset int64, limit int64) (result model.PaginationResult[T], err error) {
	opts := options.Find().SetLimit(limit).SetSkip(offset).SetSort(sort)
	cursor, err := m.coll.Find(context.Background(), filter, opts)
	cnt, _ := m.coll.CountDocuments(context.Background(), filter)
	defer cursor.Close(context.Background())
	result.Results = make([]T, 0)
	if err != nil {
		return result, err
	}
	var results []T
	if err = cursor.All(context.Background(), &results); err != nil {
		return result, err
	}
	result.Results = results
	result.Total = cnt
	result.Offset = offset
	result.Count = int64(len(results))
	return result, nil
}

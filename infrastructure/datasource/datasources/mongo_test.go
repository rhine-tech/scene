package datasources

import (
	"testing"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	loggerRepo "github.com/rhine-tech/scene/infrastructure/logger/repository"
	"github.com/stretchr/testify/require"
)

func TestMongoDataSourceSetupRejectsUnreachableDatabase(t *testing.T) {
	ds := NewMongoDataSource(datasource.MongoConfig{
		Host:     "127.0.0.1",
		Port:     1,
		Database: "scene",
		Options:  "serverSelectionTimeoutMS=10",
	})
	impl := ds.(*MongoRepo)
	impl.log = &loggerRepo.DummyLogger{}

	require.Error(t, ds.Setup())
	require.Error(t, ds.Status())
	require.NoError(t, ds.Dispose())
}

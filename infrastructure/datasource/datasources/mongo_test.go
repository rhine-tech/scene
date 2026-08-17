package datasources

import (
	"testing"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
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
	impl.log = logger.NoopLogger{}

	require.Error(t, ds.Setup())
	require.Error(t, ds.Status())
	require.NoError(t, ds.TearDown())
}

package datasources

import (
	"testing"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	loggerRepo "github.com/rhine-tech/scene/infrastructure/logger/repository"
	"github.com/stretchr/testify/require"
)

func TestPostgresDataSourceSetupRejectsUnreachableDatabase(t *testing.T) {
	ds := NewPostgresDataSource(datasource.PostgresConfig{
		Host:     "127.0.0.1",
		Port:     1,
		Username: "scene",
		Password: "secret",
		Database: "scene",
		Options:  "sslmode=disable&connect_timeout=1",
	})
	impl := ds.(*postgresImpl)
	impl.log = &loggerRepo.DummyLogger{}

	require.Error(t, ds.Setup())
	require.Error(t, ds.Status())
	require.Equal(t, "PostgresDataSource", ds.DataSourceName().Interface)
	require.NoError(t, ds.Dispose())
}

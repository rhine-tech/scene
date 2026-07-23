package orm

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/infrastructure/logger"
	loggerrepo "github.com/rhine-tech/scene/infrastructure/logger/repository"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var registerTestLogger sync.Once

func newTestGorm(t *testing.T) *Gorm {
	t.Helper()

	registerTestLogger.Do(func() {
		registry.Register[logger.ILogger](loggerrepo.NewZapColoredLogger())
	})

	dsnName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	ds := datasources.SqliteDatasource(datasource.SqliteConfig{
		Path:    "file:" + dsnName,
		Options: "mode=memory&cache=shared",
	})
	registry.Inject(ds)
	require.NoError(t, ds.Setup())
	ds.Connection().SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, ds.Dispose())
	})

	db := NewGormWithSQLite(ds)
	registry.Inject(db)
	require.NoError(t, db.Setup())
	return db
}

type transactionAccountRow struct {
	ID string `gorm:"primaryKey"`
}

type transactionAuditRow struct {
	ID string `gorm:"primaryKey"`
}

type postgresDataSourceStub struct {
	db *sql.DB
}

func (p *postgresDataSourceStub) Setup() error {
	return nil
}

func (p *postgresDataSourceStub) Dispose() error {
	return p.db.Close()
}

func (p *postgresDataSourceStub) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("PostgresDataSource")
}

func (p *postgresDataSourceStub) Status() error {
	return nil
}

func (p *postgresDataSourceStub) Connection() *sql.DB {
	return p.db
}

func TestNewGormWithPostgreSQL(t *testing.T) {
	sqlDB, err := sql.Open("pgx", datasource.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "scene",
		Options:  "sslmode=disable",
	}.DSN())
	require.NoError(t, err)

	ds := &postgresDataSourceStub{db: sqlDB}
	t.Cleanup(func() {
		require.NoError(t, ds.Dispose())
	})

	db := NewGormWithPostgreSQL(ds)
	dialector, ok := db.dialector().(*gormPostgres.Dialector)
	require.True(t, ok)
	require.Same(t, ds.Connection(), dialector.Conn)
	require.Same(t, ds, db.ds)
}

func TestGormSessionAndTransaction(t *testing.T) {
	db := newTestGorm(t)
	require.NoError(t, db.AutoMigrate(&transactionAccountRow{}, &transactionAuditRow{}))

	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "request")
	require.Equal(t, "request", db.Session(ctx).Statement.Context.Value(contextKey{}))

	rollback := errors.New("rollback")
	err := db.Transaction(ctx, func(tx *gorm.DB) error {
		require.NoError(t, tx.Create(&transactionAccountRow{ID: "account"}).Error)
		require.NoError(t, tx.Create(&transactionAuditRow{ID: "audit"}).Error)
		return rollback
	})
	require.ErrorIs(t, err, rollback)

	var accountCount int64
	require.NoError(t, db.Session(ctx).Model(&transactionAccountRow{}).Count(&accountCount).Error)
	require.Zero(t, accountCount)

	var auditCount int64
	require.NoError(t, db.Session(ctx).Model(&transactionAuditRow{}).Count(&auditCount).Error)
	require.Zero(t, auditCount)
}

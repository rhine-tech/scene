package orm

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/datasource/datasources"
	"github.com/rhine-tech/scene/infrastructure/logger"
	loggerrepo "github.com/rhine-tech/scene/infrastructure/logger/repository"
	"github.com/rhine-tech/scene/registry"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var registerTestLogger sync.Once

func newTestGorm(t *testing.T) *Gorm {
	t.Helper()

	registerTestLogger.Do(func() {
		registry.Register[logger.ILogger](loggerrepo.NewZapColoredLogger())
	})

	dsnName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	ds := datasources.SqliteDatasource(datasource.DatabaseConfig{
		Host:    "file:" + dsnName,
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

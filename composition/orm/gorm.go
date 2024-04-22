package orm

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"time"

	gormSqlite "github.com/glebarez/sqlite"
	gormMysql "gorm.io/driver/mysql"
)

type Gorm interface {
	ORM
	DB() *gorm.DB
	RegisterModel(model ...any) error
}

func GormWithMysql(ds datasource.MysqlDataSource) Gorm {
	return &gormImpl{dialector: func() gorm.Dialector {
		return gormMysql.New(gormMysql.Config{
			Conn: ds.Connection(),
		})
	}, ds: ds}
}

func GormWithSqlite(ds datasource.SqliteDataSource) Gorm {
	return &gormImpl{dialector: func() gorm.Dialector {
		return &gormSqlite.Dialector{
			Conn: ds.Connection(),
		}
	}, ds: ds}
}

type gormLogger struct {
	prefix string
	log    logger.ILogger `aperture:""`
}

func (g *gormLogger) LogMode(level gormlog.LogLevel) gormlog.Interface {
	return g
}

func (g *gormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	g.log.Infof(g.prefix+s, i...)
}

func (g *gormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	g.log.Warnf(g.prefix+s, i...)
}

func (g *gormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	g.log.Errorf(g.prefix+s, i...)
}

func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sqlStr, _ := fc()
	g.log.Debugf("trace sql: %s", sqlStr)
}

type gormImpl struct {
	db        *gorm.DB
	dialector func() gorm.Dialector
	ds        datasource.DataSource
	log       logger.ILogger `aperture:""`
}

func (g *gormImpl) OrmName() scene.ImplName {
	return Lens.ImplNameNoVer("Gorm")
}

func (g *gormImpl) Setup() error {
	g.log = g.log.WithPrefix(g.OrmName().Identifier())
	g.log.Infof("setup gorm with datasource %s", g.ds.DataSourceName().Implementation)
	gormDb, err := gorm.Open(g.dialector(), &gorm.Config{
		Logger: &gormLogger{prefix: "GormInternal: ", log: g.log},
	})
	if err != nil {
		g.log.ErrorW("create gorm instance failed", "error", err)
		return err
	}
	g.db = gormDb
	return nil
}

func (g *gormImpl) DB() *gorm.DB {
	return g.db
}

func (g *gormImpl) RegisterModel(model ...any) error {
	err := g.db.AutoMigrate(model...)
	if err != nil {
		g.log.ErrorW("register model failed when migrating model", "error", err)
		return err
	}
	g.log.Infof("register %d model success", len(model))
	return nil
}

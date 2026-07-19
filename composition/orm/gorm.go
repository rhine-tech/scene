package orm

import (
	"context"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

const Lens scene.CompositionName = "orm"

// Gorm owns the configured GORM database handle.
//
// Gorm is deliberately a concrete type. Database substitutability belongs at
// the module repository boundary; repositories using GORM should depend on
// *Gorm and use Session to build native GORM queries.
type Gorm struct {
	db        *gorm.DB
	dialector func() gorm.Dialector
	ds        datasource.DataSource
	log       logger.ILogger `aperture:""`
}

var _ scene.Named = (*Gorm)(nil)

// NewGorm creates a GORM component. The dialector factory is evaluated during
// Setup, after the injected data source has completed its own setup.
func NewGorm(dialector func() gorm.Dialector, ds datasource.DataSource) *Gorm {
	return &Gorm{
		dialector: dialector,
		ds:        ds,
	}
}

func (g *Gorm) ImplName() scene.ImplName {
	return Lens.ImplNameNoVer("Gorm")
}

// Setup opens the GORM handle over the configured data source.
func (g *Gorm) Setup() error {
	g.log.Infof("setup gorm with datasource %s", g.ds.DataSourceName().Interface)

	db, err := gorm.Open(g.dialector(), &gorm.Config{
		Logger:         &gormLogger{prefix: "GormInternal: ", log: g.log},
		TranslateError: true,
	})
	if err != nil {
		g.log.ErrorW("create gorm instance failed", "error", err)
		return err
	}
	g.db = db
	return nil
}

// Session returns a native GORM session carrying ctx.
//
// Callers should obtain a new session for each repository operation instead
// of retaining the returned *gorm.DB.
func (g *Gorm) Session(ctx context.Context) *gorm.DB {
	return g.db.WithContext(ctx)
}

// Transaction executes fn in a GORM transaction carrying ctx.
//
// Repositories that need to update multiple models can use the supplied
// *gorm.DB for every operation in the transaction.
func (g *Gorm) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return g.Session(ctx).Transaction(fn)
}

// AutoMigrate applies GORM's model migration.
//
// This is a concrete startup capability rather than a separate framework
// interface. GORM repository adapters normally call it from Setup.
func (g *Gorm) AutoMigrate(models ...any) error {
	if err := g.db.AutoMigrate(models...); err != nil {
		g.log.ErrorW("auto migrate models failed", "models", len(models), "error", err)
		return err
	}
	g.log.Infof("auto migrated %d models", len(models))
	return nil
}

type gormLogger struct {
	prefix string
	log    logger.ILogger
}

func (g *gormLogger) LogMode(gormlog.LogLevel) gormlog.Interface {
	return g
}

func (g *gormLogger) Info(_ context.Context, message string, args ...any) {
	g.log.Infof(g.prefix+message, args...)
}

func (g *gormLogger) Warn(_ context.Context, message string, args ...any) {
	g.log.Warnf(g.prefix+message, args...)
}

func (g *gormLogger) Error(_ context.Context, message string, args ...any) {
	g.log.Errorf(g.prefix+message, args...)
}

func (g *gormLogger) Trace(
	_ context.Context,
	_ time.Time,
	sql func() (statement string, rowsAffected int64),
	_ error,
) {
	statement, _ := sql()
	g.log.Debugf("trace sql: %s", statement)
}

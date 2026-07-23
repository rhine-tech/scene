package datasources

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
)

type postgresImpl struct {
	db  *sql.DB
	cfg datasource.PostgresConfig
	err error
	log logger.ILogger `aperture:""`
}

var _ datasource.PostgresDataSource = (*postgresImpl)(nil)

func NewPostgresDataSource(cfg datasource.PostgresConfig) datasource.PostgresDataSource {
	return &postgresImpl{cfg: cfg}
}

func (p *postgresImpl) Dispose() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

func (p *postgresImpl) Setup() error {
	p.log = p.log.WithPrefix(p.DataSourceName().String())
	p.db, p.err = sql.Open("pgx", p.cfg.DSN())
	if p.err != nil {
		p.log.Errorf("failed to open PostgreSQL datasource: %s", p.err)
		return p.err
	}
	if p.err = p.Status(); p.err != nil {
		_ = p.db.Close()
		p.db = nil
		p.log.Errorf(
			"failed to connect PostgreSQL datasource %s:%d/%s: %s",
			p.cfg.Host,
			p.cfg.Port,
			p.cfg.Database,
			p.err,
		)
		return p.err
	}
	p.log.Infof(
		"initialized PostgreSQL datasource %s:%d/%s",
		p.cfg.Host,
		p.cfg.Port,
		p.cfg.Database,
	)
	return nil
}

func (p *postgresImpl) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("PostgresDataSource")
}

func (p *postgresImpl) Status() error {
	return sqlDataSourceStatus(p.db, p.err)
}

func (p *postgresImpl) Connection() *sql.DB {
	return p.db
}

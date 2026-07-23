package datasources

import (
	"database/sql"
	"errors"

	_ "github.com/glebarez/go-sqlite"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
)

type sqliteImpl struct {
	db  *sql.DB
	err error
	cfg datasource.SqliteConfig
	log logger.ILogger `aperture:""`
}

func SqliteDatasource(cfg datasource.SqliteConfig) datasource.SqliteDataSource {
	return &sqliteImpl{
		cfg: cfg,
	}
}

func (s *sqliteImpl) Dispose() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *sqliteImpl) Setup() error {
	s.log = s.log.WithPrefix(s.DataSourceName().String())
	if s.cfg.DSN() == "" {
		s.log.Errorf("invalid sqlite dsn: sqlite dsn is empty")
		s.err = errors.New("invalid sqlite dsn")
		return s.err
	}
	db, err := sql.Open("sqlite", s.cfg.DSN())
	if err != nil {
		s.log.Errorf("\"%s\" failed to open: %s", s.cfg.DSN(), err)
		s.err = err
		return err
	}
	s.db = db
	s.err = nil
	if s.err = s.Status(); s.err != nil {
		_ = s.db.Close()
		s.db = nil
		s.log.Errorf("\"%s\" failed to connect: %s", s.cfg.DSN(), s.err)
		return s.err
	}
	s.log.Infof("establish connection to \"%s\" succeed", s.cfg.DSN())
	return nil
}

func (s *sqliteImpl) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("SqliteDataSource")
}

func (s *sqliteImpl) Status() error {
	return sqlDataSourceStatus(s.db, s.err)
}

func (s *sqliteImpl) Connection() *sql.DB {
	return s.db
}

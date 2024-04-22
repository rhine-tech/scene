package datasources

import (
	"database/sql"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/model"

	_ "github.com/glebarez/go-sqlite"
)

type sqliteImpl struct {
	db  *sql.DB
	err error
	cfg model.DatabaseConfig
	log logger.ILogger `aperture:""`
}

func SqliteDatasource(cfg model.DatabaseConfig) datasource.SqliteDataSource {
	return &sqliteImpl{
		cfg: cfg,
	}
}

func (s *sqliteImpl) Dispose() error {
	return s.db.Close()
}

func (s *sqliteImpl) Setup() error {
	s.log = s.log.WithPrefix(s.DataSourceName().String())
	db, err := sql.Open("sqlite", s.cfg.SqliteDSN())
	if err != nil {
		s.log.Errorf("\"%s\" failed to open: %s", s.cfg.SqliteDSN(), err)
		s.err = err
		return err
	}
	s.db = db
	s.err = nil
	s.log.Infof("establish connection to \"%s\" succeed", s.cfg.SqliteDSN())
	return nil
}

func (s *sqliteImpl) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("SqliteDataSource")
}

func (s *sqliteImpl) Status() error {
	return s.err
}

func (s *sqliteImpl) Connection() *sql.DB {
	return s.db
}

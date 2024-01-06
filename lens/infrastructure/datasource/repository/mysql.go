package repository

import (
	"database/sql"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/model"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlRepo struct {
	db  *sql.DB
	cfg model.DatabaseConfig
	err error
	log logger.ILogger `aperture:""`
}

func NewMysqlDatasource(cfg model.DatabaseConfig) datasource.MysqlDataSource {
	return &MysqlRepo{
		cfg: cfg,
	}
}

func (m *MysqlRepo) Dispose() error {
	return m.db.Close()
}

func (m *MysqlRepo) Setup() error {
	m.log = m.log.WithPrefix(m.DataSourceName().String())
	if m.err != nil {
		m.log.Errorf("\"%s\" init failed: %s", m.cfg.MysqlDSN(), m.err)
		return m.err
	}
	m.db, m.err = sql.Open("mysql", m.cfg.MysqlDSN())
	if m.err != nil {
		m.log.Errorf("\"%s\" failed to open: %s", m.cfg.MysqlDSN(), m.err)
		return m.err
	}
	m.log.Infof("establish connection to \"%s\" succeed", m.cfg.MysqlDSN())
	return nil
}

func (m *MysqlRepo) DataSourceName() scene.ImplName {
	return scene.NewRepoImplNameNoVer("datasource", "mysql")
}

func (m *MysqlRepo) Status() error {
	return m.err
}

func (m *MysqlRepo) Connection() *sql.DB {
	return m.db
}

package datasources

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/infrastructure/logger"
)

type MysqlRepo struct {
	db  *sql.DB
	cfg datasource.DatabaseConfig
	err error
	log logger.ILogger `aperture:""`
}

func NewMysqlDatasource(cfg datasource.DatabaseConfig) datasource.MysqlDataSource {
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
	return datasource.Lens.ImplNameNoVer("MysqlDataSource")
}

func (m *MysqlRepo) Status() error {
	return m.err
}

func (m *MysqlRepo) Connection() *sql.DB {
	return m.db
}

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
	cfg datasource.MysqlConfig
	err error
	log logger.ILogger `aperture:""`
}

func NewMysqlDatasource(cfg datasource.MysqlConfig) datasource.MysqlDataSource {
	return &MysqlRepo{
		cfg: cfg,
	}
}

func (m *MysqlRepo) Dispose() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}

func (m *MysqlRepo) Setup() error {
	m.log = m.log.WithPrefix(m.DataSourceName().String())
	if m.err != nil {
		m.log.Errorf("\"%s\" init failed: %s", m.cfg.MaskedDSN(), m.err)
		return m.err
	}
	m.db, m.err = sql.Open("mysql", m.cfg.DSN())
	if m.err != nil {
		m.log.Errorf("\"%s\" failed to open: %s", m.cfg.MaskedDSN(), m.err)
		return m.err
	}
	if m.err = m.Status(); m.err != nil {
		_ = m.db.Close()
		m.db = nil
		m.log.Errorf("\"%s\" failed to connect: %s", m.cfg.MaskedDSN(), m.err)
		return m.err
	}
	m.log.Infof("establish connection to \"%s\" succeed", m.cfg.MaskedDSN())
	return nil
}

func (m *MysqlRepo) DataSourceName() scene.ImplName {
	return datasource.Lens.ImplNameNoVer("MysqlDataSource")
}

func (m *MysqlRepo) Status() error {
	return sqlDataSourceStatus(m.db, m.err)
}

func (m *MysqlRepo) Connection() *sql.DB {
	return m.db
}

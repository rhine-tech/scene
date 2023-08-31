package repository

import (
	"database/sql"
	"github.com/aynakeya/scene/errcode"
	"github.com/aynakeya/scene/lens/infrastructure/datasource"
	"github.com/aynakeya/scene/model"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlRepo struct {
	db  *sql.DB
	err error
}

func NewMysqlRepo(cfg model.DatabaseConfig) datasource.MysqlDataSource {
	db, err := sql.Open("mysql", cfg.MysqlDSN())
	repo := &MysqlRepo{
		db:  db,
		err: err,
	}
	if err != nil {
		return repo
	}
	return repo
}

func (m *MysqlRepo) Dispose() error {
	return m.db.Close()
}

func (m *MysqlRepo) Setup() error {
	if m.err != nil {
		return errcode.RepositoryInitError.WithDetail(m.err)
	}
	return nil
}

func (m *MysqlRepo) DataSourceName() string {
	return "datasource.repository.mysql"
}

func (m *MysqlRepo) Status() error {
	return m.err
}

func (m *MysqlRepo) Connection() *sql.DB {
	return m.db
}

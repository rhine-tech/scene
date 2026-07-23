package datasources

import (
	"database/sql"
	"errors"
)

var errSQLDataSourceNotSetup = errors.New("sql datasource is not set up")

func sqlDataSourceStatus(db *sql.DB, setupErr error) error {
	if setupErr != nil {
		return setupErr
	}
	if db == nil {
		return errSQLDataSourceNotSetup
	}
	return db.Ping()
}

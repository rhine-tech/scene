package datasource

import "database/sql"

type SqlDataSource interface {
	DataSource
	Connection() *sql.DB
}

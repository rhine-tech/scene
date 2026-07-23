package datasource

type SqliteConfig struct {
	Path    string
	Options string
}

func (c SqliteConfig) DSN() string {
	dsn := c.Path
	if c.Options != "" {
		dsn += "?" + c.Options
	}
	return dsn
}

type SqliteDataSource interface {
	SqlDataSource
}

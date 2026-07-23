package datasource

import (
	"net"
	"net/url"
	"strconv"
)

type PostgresConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Options  string
}

func (c PostgresConfig) DSN() string {
	uri := url.URL{
		Scheme:   "postgres",
		Host:     net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:     c.Database,
		RawQuery: c.Options,
	}
	if c.Username != "" {
		if c.Password == "" {
			uri.User = url.User(c.Username)
		} else {
			uri.User = url.UserPassword(c.Username, c.Password)
		}
	}
	return uri.String()
}

type PostgresDataSource interface {
	SqlDataSource
}

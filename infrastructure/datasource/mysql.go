package datasource

import (
	"fmt"
	"strings"
)

type MysqlConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Options  string
}

func (c MysqlConfig) DSN() string {
	var sb strings.Builder
	if c.Username != "" {
		sb.WriteString(fmt.Sprintf("%s:%s@", c.Username, c.Password))
	}
	sb.WriteString(fmt.Sprintf("tcp(%s:%d)", c.Host, c.Port))
	sb.WriteString(fmt.Sprintf("/%s", c.Database))
	if c.Options != "" {
		sb.WriteString(fmt.Sprintf("?%s", c.Options))
	}
	return sb.String()
}

func (c MysqlConfig) MaskedDSN() string {
	if c.Password != "" {
		c.Password = maskedPassword
	}
	return c.DSN()
}

type MysqlDataSource interface {
	SqlDataSource
}

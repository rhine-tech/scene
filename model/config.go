package model

import (
	"fmt"
	"strings"
)

type FileConfig struct {
	Path string
}

func NewFileConfig(path string) FileConfig {
	return FileConfig{
		Path: path,
	}
}

// DatabaseConfig is a struct that contains the configuration for a database.
// Universal database configuration.
// "scfg" is the tag name used in infrastructure/config.
type DatabaseConfig struct {
	Host     string `scfg:"host"`
	Port     int    `scfg:"port"`
	Username string `scfg:"username"`
	Password string `scfg:"password"`
	Database string `scfg:"database,default=scene"`
	Options  string `scfg:"options"` // in format of "key1=value1&key2=value2"
}

// DatabaseConfigFromDSN parses a DSN string and returns a DatabaseConfig.
func DatabaseConfigFromDSN(dsn string) DatabaseConfig {
	var db DatabaseConfig
	return db
}

func (d DatabaseConfig) MongoDSN() string {
	var uri string
	if d.Username == "" {
		uri = fmt.Sprintf("mongodb://%s:%d/", d.Host, d.Port)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/",
			d.Username, d.Password, d.Host, d.Port)
	}
	if d.Options != "" {
		uri += "?" + d.Options
	}
	return uri
}

func (d DatabaseConfig) MysqlDSN() string {
	var sb strings.Builder
	if d.Username != "" {
		sb.WriteString(fmt.Sprintf("%s:%s@", d.Username, d.Password))
	}
	sb.WriteString(fmt.Sprintf("tcp(%s:%d)", d.Host, d.Port))
	sb.WriteString(fmt.Sprintf("/%s", d.Database))
	if d.Options != "" {
		sb.WriteString(fmt.Sprintf("?%s", d.Options))
	}
	return sb.String()
}

func (d DatabaseConfig) RedisDSN() string {
	var sb strings.Builder
	sb.WriteString("redis://")
	if d.Username != "" {
		sb.WriteString(fmt.Sprintf("%s:%s@", d.Username, d.Password))
	}
	sb.WriteString(fmt.Sprintf("%s:%d/%s", d.Host, d.Port, d.Database))
	return sb.String()
}

func (d DatabaseConfig) SqliteDSN() string {
	val := d.Host
	if d.Options != "" {
		val += "?" + d.Options
	}
	return val
}

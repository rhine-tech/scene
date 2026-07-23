package datasource

import (
	"net"
	"net/url"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoConfig struct {
	Host          string
	Port          int
	Username      string
	Password      string
	Database      string
	AuthSource    string
	Options       string
	UseAPIVersion bool
}

func (c MongoConfig) DSN() string {
	uri := url.URL{
		Scheme: "mongodb",
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   "/",
	}
	if c.Username != "" {
		uri.User = url.UserPassword(c.Username, c.Password)
	}

	options := c.Options
	if c.AuthSource != "" {
		if options != "" {
			options += "&"
		}
		options += "authSource=" + url.QueryEscape(c.AuthSource)
	}
	uri.RawQuery = options
	return uri.String()
}

func (c MongoConfig) MaskedDSN() string {
	if c.Password != "" {
		c.Password = maskedPassword
	}
	return c.DSN()
}

type MongoDataSource interface {
	DataSource
	Database() *mongo.Database
	Collection(coll string) *mongo.Collection
}

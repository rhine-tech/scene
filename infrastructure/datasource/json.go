package datasource

type JsonConfig struct {
	Path string
}

type JsonDataSource interface {
	DataSource
	Load() ([]byte, error)
	Save(data []byte) error
}

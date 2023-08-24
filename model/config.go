package model

type FileConfig struct {
	Path string
}

func NewFileConfig(path string) FileConfig {
	return FileConfig{
		Path: path,
	}
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

package config

const TagName = "scfg"

type ConfigProvider interface {
	Init() error
	GetString(key string) string
	GetInt(key string) int64
	GetBool(key string) bool
	GetStringE(key string) (string, bool)
	GetIntE(key string) (int64, bool)
	GetBoolE(key string) (bool, bool)
}

type ConfigUnmarshaler interface {
	ConfigProvider
	Unmarshal(val interface{}) error
	UnmarshalWithPrefix(prefix string, val interface{}) error
}

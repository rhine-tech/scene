package repository

import (
	"os"

	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/spf13/cast"
)

type envCfg struct{}

func NewEnvMarshaller() config.IConfig {
	return config.NewConfigUnmarshaler(&envCfg{})
}

func (cfg *envCfg) Init() error {
	return nil
}

func (cfg *envCfg) GetStringE(key string) (string, bool) {
	key = toUnderscoreKey(key)
	if v, ok := os.LookupEnv(key); ok {
		return v, true
	}
	return "", false
}

func (cfg *envCfg) GetIntE(key string) (int64, bool) {
	key = toUnderscoreKey(key)
	if v, ok := os.LookupEnv(key); ok {
		val, err := cast.ToInt64E(v)
		return val, err == nil
	}
	return 0, false
}

func (cfg *envCfg) GetBoolE(key string) (bool, bool) {
	key = toUnderscoreKey(key)
	if v, ok := os.LookupEnv(key); ok {
		val, err := cast.ToBoolE(v)
		return val, err == nil
	}
	return false, false
}

func (cfg *envCfg) GetString(key string) string {
	val, _ := cfg.GetStringE(key)
	return val
}

func (cfg *envCfg) GetInt(key string) int64 {
	val, _ := cfg.GetIntE(key)
	return val
}

func (cfg *envCfg) GetBool(key string) bool {
	val, _ := cfg.GetBoolE(key)
	return val
}

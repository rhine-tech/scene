package cfgur

import (
	"github.com/spf13/cast"
	"os"
)

type envCfg struct {
}

func (cfg *envCfg) Init() error {
	return nil
}

func NewEnvMarshaller() ConfigUnmarshaler {
	return &commonMarshaller{ConfigProvider: &envCfg{}}
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
		return cast.ToInt64(v), true
	}
	return 0, false
}

func (cfg *envCfg) GetBoolE(key string) (bool, bool) {
	key = toUnderscoreKey(key)
	if v, ok := os.LookupEnv(key); ok {
		return cast.ToBool(v), true
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

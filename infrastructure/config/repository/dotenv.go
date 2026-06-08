package repository

import (
	"github.com/joho/godotenv"
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/spf13/cast"
)

type dotenvCfg struct {
	filenames []string
	envs      map[string]string
}

func NewDotenvMarshaller(filenames ...string) config.IConfig {
	return config.NewConfigUnmarshaler(&dotenvCfg{filenames: filenames})
}

func (cfg *dotenvCfg) Init() error {
	envs, err := godotenv.Read(cfg.filenames...)
	if err != nil {
		return err
	}
	cfg.envs = envs
	return nil
}

func (cfg *dotenvCfg) GetStringE(key string) (string, bool) {
	key = toUnderscoreKey(key)
	if v, ok := cfg.envs[key]; ok {
		return v, true
	}
	return "", false
}

func (cfg *dotenvCfg) GetIntE(key string) (int64, bool) {
	key = toUnderscoreKey(key)
	if v, ok := cfg.envs[key]; ok {
		val, err := cast.ToInt64E(v)
		return val, err == nil
	}
	return 0, false
}

func (cfg *dotenvCfg) GetBoolE(key string) (bool, bool) {
	key = toUnderscoreKey(key)
	if v, ok := cfg.envs[key]; ok {
		val, err := cast.ToBoolE(v)
		return val, err == nil
	}
	return false, false
}

func (cfg *dotenvCfg) GetString(key string) string {
	val, _ := cfg.GetStringE(key)
	return val
}

func (cfg *dotenvCfg) GetInt(key string) int64 {
	val, _ := cfg.GetIntE(key)
	return val
}

func (cfg *dotenvCfg) GetBool(key string) bool {
	val, _ := cfg.GetBoolE(key)
	return val
}

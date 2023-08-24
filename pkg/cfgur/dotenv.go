package cfgur

import (
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

type dotenvCfg struct {
	filenames []string
	envs      map[string]string
}

func (cfg *dotenvCfg) Init() error {
	envs, err := godotenv.Read(cfg.filenames...)
	if err != nil {
		return err
	}
	cfg.envs = envs
	return nil
}

func NewDotenvMarshaller(filenames ...string) ConfigUnmarshaler {
	return &commonMarshaller{ConfigProvider: &dotenvCfg{filenames: filenames}}
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
		return cast.ToInt64(v), true
	}
	return 0, false
}

func (cfg *dotenvCfg) GetBoolE(key string) (bool, bool) {
	key = toUnderscoreKey(key)
	if v, ok := cfg.envs[key]; ok {
		return cast.ToBool(v), true
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

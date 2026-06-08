package repository

import (
	"os"
	"strconv"

	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/tidwall/gjson"
)

type jsonCfg struct {
	filepath string
	val      gjson.Result
}

func NewJsonMarshaller(filepath string) config.IConfig {
	return config.NewConfigUnmarshaler(&jsonCfg{filepath: filepath})
}

func (cfg *jsonCfg) Init() error {
	bytes, err := os.ReadFile(cfg.filepath)
	if err != nil {
		return err
	}
	cfg.val = gjson.ParseBytes(bytes)
	return nil
}

func (cfg *jsonCfg) GetStringE(key string) (string, bool) {
	result := cfg.val.Get(key)
	if result.Exists() {
		return result.String(), true
	}
	return "", false
}

func (cfg *jsonCfg) GetIntE(key string) (int64, bool) {
	result := cfg.val.Get(key)
	if !result.Exists() {
		return 0, false
	}
	if result.Type == gjson.Number {
		return result.Int(), true
	}
	if result.Type == gjson.String {
		v, err := strconv.ParseInt(result.String(), 10, 64)
		return v, err == nil
	}
	return 0, false
}

func (cfg *jsonCfg) GetBoolE(key string) (bool, bool) {
	result := cfg.val.Get(key)
	if !result.Exists() {
		return false, false
	}
	if result.Type == gjson.True || result.Type == gjson.False {
		return result.Bool(), true
	}
	if result.Type == gjson.String {
		v, err := strconv.ParseBool(result.String())
		return v, err == nil
	}
	return false, false
}

func (cfg *jsonCfg) GetString(key string) string {
	val, _ := cfg.GetStringE(key)
	return val
}

func (cfg *jsonCfg) GetInt(key string) int64 {
	val, _ := cfg.GetIntE(key)
	return val
}

func (cfg *jsonCfg) GetBool(key string) bool {
	val, _ := cfg.GetBoolE(key)
	return val
}

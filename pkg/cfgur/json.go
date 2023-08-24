package cfgur

import (
	"github.com/tidwall/gjson"
	"os"
)

type jsonCfg struct {
	filepath string
	val      gjson.Result
}

func (cfg *jsonCfg) Init() error {
	bytes, err := os.ReadFile(cfg.filepath)
	if err != nil {
		return err
	}
	cfg.val = gjson.ParseBytes(bytes)
	return nil
}

func NewJsonMarshaller(filepath string) ConfigUnmarshaler {
	return &commonMarshaller{ConfigProvider: &jsonCfg{filepath: filepath}}
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
	if result.Exists() {
		return result.Int(), true
	}
	return 0, false
}

func (cfg *jsonCfg) GetBoolE(key string) (bool, bool) {
	result := cfg.val.Get(key)
	if result.Exists() {
		return result.Bool(), true
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

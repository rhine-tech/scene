package cfgur

import (
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cast"
)

type tomlCfg struct {
	filename string
	data     map[string]interface{}
}

// NewTomlConfig creates a new instance of TOML configuration manager
func NewTomlConfig(filename string) ConfigUnmarshaler {
	return &commonMarshaller{ConfigProvider: &tomlCfg{
		filename: filename,
	}}
}

func (cfg *tomlCfg) Init() error {
	m := make(map[string]interface{})
	if _, err := toml.DecodeFile(cfg.filename, &m); err != nil {
		return err
	}
	cfg.data = m
	return nil
}

// get traverses nested maps using "a.b.c" style keys
func (cfg *tomlCfg) get(key string) (interface{}, bool) {
	if cfg.data == nil {
		return nil, false
	}

	var cur interface{} = cfg.data
	parts := strings.Split(key, ".")

	for _, part := range parts {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		v, ok := m[part]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func (cfg *tomlCfg) GetString(key string) string {
	v, _ := cfg.GetStringE(key)
	return v
}

func (cfg *tomlCfg) GetInt(key string) int64 {
	v, _ := cfg.GetIntE(key)
	return v
}

func (cfg *tomlCfg) GetBool(key string) bool {
	v, _ := cfg.GetBoolE(key)
	return v
}

func (cfg *tomlCfg) GetStringE(key string) (string, bool) {
	v, ok := cfg.get(key)
	if !ok {
		return "", false
	}
	s, err := cast.ToStringE(v)
	return s, err == nil
}

func (cfg *tomlCfg) GetIntE(key string) (int64, bool) {
	v, ok := cfg.get(key)
	if !ok {
		return 0, false
	}
	i, err := cast.ToInt64E(v)
	return i, err == nil
}

func (cfg *tomlCfg) GetBoolE(key string) (bool, bool) {
	v, ok := cfg.get(key)
	if !ok {
		return false, false
	}
	b, err := cast.ToBoolE(v)
	return b, err == nil
}

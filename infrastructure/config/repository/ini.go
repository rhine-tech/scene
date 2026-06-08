package repository

import (
	"strings"

	"github.com/rhine-tech/scene/infrastructure/config"
	"gopkg.in/ini.v1"
)

type iniCfg struct {
	file     *ini.File
	filename string
}

func NewIniConfig(filename string) config.IConfig {
	return config.NewConfigUnmarshaler(&iniCfg{filename: filename})
}

func (cfg *iniCfg) Init() error {
	file, err := ini.Load(cfg.filename)
	if err != nil {
		return err
	}
	cfg.file = file
	return nil
}

func (cfg *iniCfg) parseKey(key string) (string, string) {
	parts := strings.Split(key, ".")
	if len(parts) >= 2 {
		return parts[0], strings.Join(parts[1:], "_")
	}
	return "", parts[0]
}

func (cfg *iniCfg) getValue(key string) *ini.Key {
	section, key := cfg.parseKey(key)
	return cfg.file.Section(section).Key(key)
}

func (cfg *iniCfg) hasValue(key string) bool {
	section, key := cfg.parseKey(key)
	sec, err := cfg.file.GetSection(section)
	if err != nil {
		return false
	}
	return sec.HasKey(key)
}

func (cfg *iniCfg) GetString(key string) string {
	v, _ := cfg.GetStringE(key)
	return v
}

func (cfg *iniCfg) GetInt(key string) int64 {
	v, _ := cfg.GetIntE(key)
	return v
}

func (cfg *iniCfg) GetBool(key string) bool {
	v, _ := cfg.GetBoolE(key)
	return v
}

func (cfg *iniCfg) GetStringE(key string) (string, bool) {
	if !cfg.hasValue(key) {
		return "", false
	}
	return cfg.getValue(key).String(), true
}

func (cfg *iniCfg) GetIntE(key string) (int64, bool) {
	if !cfg.hasValue(key) {
		return 0, false
	}
	v, err := cfg.getValue(key).Int64()
	return v, err == nil
}

func (cfg *iniCfg) GetBoolE(key string) (bool, bool) {
	if !cfg.hasValue(key) {
		return false, false
	}
	v, err := cfg.getValue(key).Bool()
	return v, err == nil
}

package cfgur

import (
	"gopkg.in/ini.v1"
	"strings"
)

// iniCfg holds the INI file data
type iniCfg struct {
	file     *ini.File
	filename string
}

// NewIniConfig creates a new instance of INI configuration manager
func NewIniConfig(filename string) ConfigUnmarshaler {
	return &commonMarshaller{ConfigProvider: &iniCfg{
		filename: filename,
	}}
}

// parseKey splits the key into section and key
func (cfg *iniCfg) parseKey(key string) (string, string) {
	parts := strings.Split(key, ".")
	if len(parts) >= 2 {
		return parts[0], strings.Join(parts[1:], "_")
	}
	return "", parts[0]
}

// getValue is a helper function to get the value from the INI file
func (cfg *iniCfg) getValue(key string) *ini.Key {
	section, key := cfg.parseKey(key)
	return cfg.file.Section(section).Key(key)
}

func (cfg *iniCfg) Init() error {
	file, err := ini.Load(cfg.filename)
	if err != nil {
		return err
	}
	cfg.file = file
	return nil
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
	return cfg.getValue(key).String(), true
}

func (cfg *iniCfg) GetIntE(key string) (int64, bool) {
	v, err := cfg.getValue(key).Int64()
	return v, err == nil
}

func (cfg *iniCfg) GetBoolE(key string) (bool, bool) {
	v, err := cfg.getValue(key).Bool()
	return v, err == nil
}

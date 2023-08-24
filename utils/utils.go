package utils

import "github.com/spf13/viper"

func CreateViper(path string) *viper.Viper {
	v := viper.New()
	v.SetConfigFile(path)
	_ = v.ReadInConfig()
	return v
}

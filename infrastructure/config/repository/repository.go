package repository

import (
	"github.com/rhine-tech/scene/infrastructure/config"
)

func NewDotEnvironmentCfgur(filenames ...string) config.IConfig {
	return NewDotenvMarshaller(filenames...)
}

func NewJsonCfgur(filename string) config.IConfig {
	return NewJsonMarshaller(filename)
}

func NewEnvironmentCfgur() config.IConfig {
	return NewEnvMarshaller()
}

func NewINICfgur(filename string) config.IConfig {
	return NewIniConfig(filename)
}

func NewTomlCfgur(filename string) config.IConfig {
	return NewTomlConfig(filename)
}

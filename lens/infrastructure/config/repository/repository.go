package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/config"
	"github.com/rhine-tech/scene/pkg/cfgur"
)

func NewDotEnvironmentCfgur(filenames ...string) config.ConfigUnmarshaler {
	return cfgur.NewDotenvMarshaller(filenames...)
}

func NewJsonCfgur(filename string) config.ConfigUnmarshaler {
	return cfgur.NewJsonMarshaller(filename)
}

func NewEnvironmentCfgur() config.ConfigUnmarshaler {
	return cfgur.NewEnvMarshaller()
}

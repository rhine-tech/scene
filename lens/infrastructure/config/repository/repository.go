package repository

import (
	"github.com/aynakeya/scene/lens/infrastructure/config"
	"github.com/aynakeya/scene/pkg/cfgur"
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

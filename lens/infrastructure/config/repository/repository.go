package repository

import (
	"github.com/rhine-tech/scene/lens/infrastructure/config"
	"github.com/rhine-tech/scene/pkg/cfgur"
)

func NewDotEnvironmentCfgur(filenames ...string) config.ConfigUnmarshaler {
	return &commonMarshaller{cfgur.NewDotenvMarshaller(filenames...)}
}

func NewJsonCfgur(filename string) config.ConfigUnmarshaler {
	return &commonMarshaller{cfgur.NewJsonMarshaller(filename)}
}

func NewEnvironmentCfgur() config.ConfigUnmarshaler {
	return &commonMarshaller{cfgur.NewEnvMarshaller()}
}

func NewINICfgur(filename string) config.ConfigUnmarshaler {
	return &commonMarshaller{cfgur.NewIniConfig(filename)}
}

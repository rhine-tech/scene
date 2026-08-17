package factory

import (
	"github.com/rhine-tech/scene/infrastructure/config"
	"github.com/rhine-tech/scene/infrastructure/config/repository"
	"github.com/rhine-tech/scene/registry"
)

func Init(configFile string) {
	setConfig(repository.NewDotEnvironmentCfgur(configFile))
}

func InitDotEnv(configFile string) {
	setConfig(repository.NewDotEnvironmentCfgur(configFile))
}

func InitJson(configFile string) {
	setConfig(repository.NewJsonCfgur(configFile))
}

func InitEnv() {
	setConfig(repository.NewEnvironmentCfgur())
}

func InitINI(configFile string) {
	setConfig(repository.NewINICfgur(configFile))
}

func InitToml(configFile string) {
	setConfig(repository.NewTomlCfgur(configFile))
}

func setConfig(cfg config.IConfig) {
	if err := cfg.Init(); err != nil {
		panic(err)
	}
	registry.Set[config.IConfig](cfg)
	registry.Set[config.ConfigUnmarshaler](cfg)
	registry.Set[config.ConfigProviderWithDefault](cfg)
}

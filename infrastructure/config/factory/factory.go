package factory

import (
	"github.com/rhine-tech/scene/lens/infrastructure/config/repository"
	"github.com/rhine-tech/scene/registry"
)

func Init(configFile string) {
	registry.RegisterConfig(repository.NewDotEnvironmentCfgur(configFile))
	if err := registry.Config.Init(); err != nil {
		panic(err)
	}
}

func InitDotEnv(configFile string) {
	registry.RegisterConfig(repository.NewDotEnvironmentCfgur(configFile))
	if err := registry.Config.Init(); err != nil {
		panic(err)
	}
}

func InitJson(configFile string) {
	registry.RegisterConfig(repository.NewJsonCfgur(configFile))
	if err := registry.Config.Init(); err != nil {
		panic(err)
	}
}

func InitEnv() {
	registry.RegisterConfig(repository.NewEnvironmentCfgur())
	if err := registry.Config.Init(); err != nil {
		panic(err)
	}
}

func InitINI(configFile string) {
	registry.RegisterConfig(repository.NewINICfgur(configFile))
	if err := registry.Config.Init(); err != nil {
		panic(err)
	}
}

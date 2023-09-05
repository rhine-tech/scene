package builder

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

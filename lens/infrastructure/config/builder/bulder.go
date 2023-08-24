package builder

import (
	"github.com/aynakeya/scene/lens/infrastructure/config/repository"
	"github.com/aynakeya/scene/registry"
)

func Init(configFile string) {
	registry.RegisterConfig(repository.NewDotEnvironmentCfgur(configFile))
	if err := registry.Config.Init(); err != nil {
		panic(err)
	}
}

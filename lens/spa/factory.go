package spa

import (
	"embed"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/spa/delivery"
)

type SPA struct {
	scene.ModuleFactory
	Embed    *embed.FS
	Router   string // Name of the Gin Router that owns this SPA.
	FsPrefix string
}

func (S SPA) Apps() []scene.Application {
	return []scene.Application{
		delivery.NewGinSPA(S.Embed, S.Router, S.FsPrefix),
	}
}

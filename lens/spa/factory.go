package spa

import (
	"embed"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/spa/delivery"
)

type SPA struct {
	scene.ModuleFactory
	Embed     *embed.FS
	UrlPrefix string
	FsPrefix  string
}

func (S SPA) Apps() []scene.Application {
	return []scene.Application{
		delivery.NewGinSPA(S.Embed, S.UrlPrefix, S.FsPrefix),
	}
}

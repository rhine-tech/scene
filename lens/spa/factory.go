package spa

import (
	"embed"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/spa/delivery"
	sgin "github.com/rhine-tech/scene/scenes/gin"
)

type SPA struct {
	Embed     *embed.FS
	UrlPrefix string
	FsPrefix  string
}

func (S SPA) Init() scene.LensInit {
	return func() {
	}
}

func (S SPA) Apps() []any {
	return []any{
		func() sgin.GinApplication {
			return delivery.NewGinSPA(S.Embed, S.UrlPrefix, S.FsPrefix)
		},
	}
}

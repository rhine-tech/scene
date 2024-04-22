package orm

import "github.com/rhine-tech/scene"

const Lens scene.CompositionName = "orm"

type ORM interface {
	OrmName() scene.ImplName
}

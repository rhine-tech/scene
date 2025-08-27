package registry

import (
	"context"
	"github.com/rhine-tech/scene"
)

var EmptyContext = context.Background()

var Disposable Registry[int, scene.Disposable]
var Setupable Registry[int, scene.Setupable]

var registrants []Registrant

func init() {
	Disposable = NewOrderedRegistry(indexedNaming[scene.Disposable]())
	Setupable = NewOrderedRegistry(indexedNaming[scene.Setupable]())

	registrants = []Registrant{
		//registrantWrapper(Repository), registrantWrapper(Service),
		registrantWrapper(Disposable), registrantWrapper(Setupable),
	}
}

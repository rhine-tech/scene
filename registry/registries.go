package registry

import (
	"context"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
)

var EmptyContext = context.Background()

// var Repository Registry[string, scene.Repository]
// var Service Registry[string, scene.Service]
var Disposable Registry[int, scene.Disposable]
var Setupable Registry[int, scene.Setupable]
var DBConfig Registry[string, *model.DatabaseConfig]

var registrants []Registrant

func init() {
	//Repository = NewRegistry(func(value scene.Repository) string {
	//	return value.RepoImplName()
	//})
	//Service = NewRegistry(func(value scene.Service) string {
	//	return value.SrvImplName()
	//})
	Disposable = NewOrderedRegistry(indexedNaming[scene.Disposable]())
	Setupable = NewOrderedRegistry(indexedNaming[scene.Setupable]())

	DBConfig = NewRegistry(func(value *model.DatabaseConfig) string {
		return value.Database
	})

	registrants = []Registrant{
		//registrantWrapper(Repository), registrantWrapper(Service),
		registrantWrapper(Disposable), registrantWrapper(Setupable),
		registrantWrapper(DBConfig),
	}
}

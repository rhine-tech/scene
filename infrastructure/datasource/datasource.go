package datasource

import "github.com/rhine-tech/scene"

const Lens scene.InfraName = "datasource"

const maskedPassword = "REDACTED"

type DataSource interface {
	scene.Disposable
	scene.Setupable
	DataSourceName() scene.ImplName
	Status() error
}

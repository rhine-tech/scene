package datasource

import "github.com/rhine-tech/scene"

const Lens scene.ModuleName = "datasource"

const maskedPassword = "REDACTED"

type DataSource interface {
	scene.Lifecycle
	DataSourceName() scene.ImplName
	Status() error
}

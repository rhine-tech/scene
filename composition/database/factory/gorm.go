package factory

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/database"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/registry"
)

type GormMysql struct {
}

func (g GormMysql) Init() scene.LensInit {
	return func() {
		registry.Register[database.Gorm](database.GormWithMysql(registry.Use(datasource.MysqlDataSource(nil))))
	}
}

func (g GormMysql) Apps() []any {
	return nil
}

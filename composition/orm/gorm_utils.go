package orm

import "gorm.io/gorm/clause"

func (g *GormRepository[Model]) toClauseColumns(keys []string) []clause.Column {
	cols := make([]clause.Column, len(keys))
	for i, key := range keys {
		cols[i] = clause.Column{Name: key}
	}
	return cols
}

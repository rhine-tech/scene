package gorm

import (
	sfilter "github.com/rhine-tech/scene/model/filter"
	"gorm.io/gorm"
)

type filterBuilder struct {
	fieldMapper sfilter.FieldMapper
	basedb      *gorm.DB
}

func (b *filterBuilder) buildFieldFilter(db *gorm.DB, filter *sfilter.FieldFilter) *gorm.DB {
	fname := b.fieldMapper.Get(filter.Field)
	switch filter.Operator {
	case sfilter.OpEqual:
		return db.Where(fname+" = ?", filter.Value)
	case sfilter.OpNotEqual:
		return db.Not(fname, filter.Value)
	case sfilter.OpGreater:
		return db.Where(fname+" > ?", filter.Value)
	case sfilter.OpGreaterOrEqual:
		return db.Where(fname+" >= ?", filter.Value)
	case sfilter.OpLess:
		return db.Where(fname+" < ?", filter.Value)
	case sfilter.OpLessOrEqual:
		return db.Where(fname+" <= ?", filter.Value)
	}
	return db
}

func (b *filterBuilder) buildLogicalFilter(db *gorm.DB, filter *sfilter.LogicalFilter) *gorm.DB {
	if len(filter.Filters) == 0 {
		return db
	}
	if len(filter.Filters) == 1 && (filter.Operator == sfilter.OpAnd || filter.Operator == sfilter.OpOr) {
		return b.buildFilter(db, filter.Filters[0])
	}
	if filter.Operator == sfilter.OpNot {
		return db.Not(b.buildFilter(db, filter.Filters[0]))
	}
	if filter.Operator == sfilter.OpAnd {
		for _, f := range filter.Filters {
			db = db.Where(b.buildFilter(db, f))
		}
		return db
	}
	if filter.Operator == sfilter.OpOr {
		for _, f := range filter.Filters {
			db = db.Or(b.buildFilter(db, f))
		}
		return db
	}
	return db
}

func (b *filterBuilder) buildFilter(db *gorm.DB, filter sfilter.Filter) *gorm.DB {
	switch filter.FilterType() {
	case sfilter.FilterTypeField:
		return b.buildFieldFilter(db, filter.(*sfilter.FieldFilter))
	case sfilter.FilterTypeLogical:
		return b.buildLogicalFilter(db, filter.(*sfilter.LogicalFilter))
	}
	return db
}

func (b *filterBuilder) BuildFilter(filters ...sfilter.Filter) *gorm.DB {
	return b.buildFilter(b.basedb, sfilter.And(filters...))
}

func (b *filterBuilder) WithFieldMapper(mapper sfilter.FieldMapper) sfilter.FilterBuilder[*gorm.DB] {
	return &filterBuilder{
		fieldMapper: mapper,
		basedb:      b.basedb,
	}
}

func (g *gormImpl) BuildFilter(filters ...sfilter.Filter) *gorm.DB {
	if len(filters) == 0 {
		return g.db
	}
	return (&filterBuilder{fieldMapper: sfilter.EmptyFieldMapper, basedb: g.db}).BuildFilter(filters...)
}

func (g *gormImpl) WithFieldMapper(mapper sfilter.FieldMapper) sfilter.FilterBuilder[*gorm.DB] {
	return &filterBuilder{
		fieldMapper: mapper,
		basedb:      g.db,
	}
}

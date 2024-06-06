package gorm

import (
	sopt "github.com/rhine-tech/scene/model/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type filterBuilder struct {
	fieldMapper sopt.FieldMapper
	basedb      *gorm.DB
}

func (b *filterBuilder) buildFieldFilter(db *gorm.DB, filterOpt *sopt.Filter) *gorm.DB {
	fname := b.fieldMapper.Get(filterOpt.Field)
	switch filterOpt.Operator {
	case sopt.OpEqual:
		return db.Where(fname+" = ?", filterOpt.Value)
	case sopt.OpNotEqual:
		return db.Not(fname, filterOpt.Value)
	case sopt.OpGreater:
		return db.Where(fname+" > ?", filterOpt.Value)
	case sopt.OpGreaterOrEqual:
		return db.Where(fname+" >= ?", filterOpt.Value)
	case sopt.OpLess:
		return db.Where(fname+" < ?", filterOpt.Value)
	case sopt.OpLessOrEqual:
		return db.Where(fname+" <= ?", filterOpt.Value)
	}
	return db
}

func (b *filterBuilder) buildLogicalFilter(db *gorm.DB, logicalOpt *sopt.Logical) *gorm.DB {
	if len(logicalOpt.Filters) == 0 {
		return db
	}
	if len(logicalOpt.Filters) == 1 && (logicalOpt.Operator == sopt.OpAnd || logicalOpt.Operator == sopt.OpOr) {
		return b.buildOption(db, logicalOpt.Filters[0])
	}
	if logicalOpt.Operator == sopt.OpNot {
		return db.Not(b.buildOption(db, logicalOpt.Filters[0]))
	}
	if logicalOpt.Operator == sopt.OpAnd {
		for _, f := range logicalOpt.Filters {
			db = db.Where(b.buildOption(db, f))
		}
		return db
	}
	if logicalOpt.Operator == sopt.OpOr {
		for _, f := range logicalOpt.Filters {
			db = db.Or(b.buildOption(db, f))
		}
		return db
	}
	return db
}

func (b *filterBuilder) buildSort(db *gorm.DB, sortOpt *sopt.Order) *gorm.DB {
	fname := b.fieldMapper.Get(sortOpt.Field)
	if sortOpt.Order == sopt.Ascending {
		return db.Order(clause.OrderByColumn{Column: clause.Column{Name: fname}, Desc: false})
	}
	return db.Order(clause.OrderByColumn{Column: clause.Column{Name: fname}, Desc: true})
}

func (b *filterBuilder) buildOption(db *gorm.DB, filter sopt.Option) *gorm.DB {
	switch filter.OptionType() {
	case sopt.OptionTypeFilter:
		return b.buildFieldFilter(db, filter.(*sopt.Filter))
	case sopt.OptionTypeLogical:
		return b.buildLogicalFilter(db, filter.(*sopt.Logical))
	case sopt.OptionTypeSort:
		return b.buildSort(db, filter.(*sopt.Order))
	}
	return db
}

func (b *filterBuilder) Build(options ...sopt.Option) *gorm.DB {
	return b.buildOption(b.basedb, sopt.And(options...))
}

func (b *filterBuilder) WithFieldMapper(mapper sopt.FieldMapper) sopt.QueryBuilder[*gorm.DB] {
	return &filterBuilder{
		fieldMapper: mapper,
		basedb:      b.basedb,
	}
}

func (g *gormImpl) Build(options ...sopt.Option) *gorm.DB {
	if len(options) == 0 {
		return g.db
	}
	return (&filterBuilder{fieldMapper: sopt.EmptyFieldMapper, basedb: g.db}).Build(options...)
}

func (g *gormImpl) WithFieldMapper(mapper sopt.FieldMapper) sopt.QueryBuilder[*gorm.DB] {
	return &filterBuilder{
		fieldMapper: mapper,
		basedb:      g.db,
	}
}

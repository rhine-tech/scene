package filter

type LogicalFilter struct {
	Operator LogicalOperator
	Filters  []Filter // FieldFilter or LogicalFilter
}

func (l *LogicalFilter) FilterType() FilterType {
	return FilterTypeLogical
}

func And(filters ...Filter) *LogicalFilter {
	return &LogicalFilter{
		Operator: OpAnd,
		Filters:  filters,
	}
}

func Or(filters ...Filter) *LogicalFilter {
	return &LogicalFilter{
		Operator: OpOr,
		Filters:  filters,
	}
}

func Not(filter Filter) *LogicalFilter {
	return &LogicalFilter{
		Operator: OpNot,
		Filters:  []Filter{filter},
	}
}

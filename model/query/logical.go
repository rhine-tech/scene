package query

type LogicalOperator string

const (
	OpAnd LogicalOperator = "and"
	OpOr  LogicalOperator = "or"
	OpNot LogicalOperator = "not"
)

type Logical struct {
	Operator LogicalOperator
	Filters  []Option // Filter or Logical
}

func (l *Logical) OptionType() OptionType {
	return OptionTypeLogical
}

func And(filters ...Option) *Logical {
	return &Logical{
		Operator: OpAnd,
		Filters:  filters,
	}
}

func Or(filters ...Option) *Logical {
	return &Logical{
		Operator: OpOr,
		Filters:  filters,
	}
}

func Not(filter Option) *Logical {
	return &Logical{
		Operator: OpNot,
		Filters:  []Option{filter},
	}
}

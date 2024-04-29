package filter

type Operator string

const (
	OpEqual          Operator = "eq"
	OpNotEqual       Operator = "ne"
	OpGreater        Operator = "gt"
	OpGreaterOrEqual Operator = "gte"
	OpLess           Operator = "lt"
	OpLessOrEqual    Operator = "lte"
)

type LogicalOperator string

const (
	OpAnd LogicalOperator = "and"
	OpOr  LogicalOperator = "or"
	OpNot LogicalOperator = "not"
)

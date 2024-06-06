package query

type Operator string

const (
	OpEqual          Operator = "eq"
	OpNotEqual       Operator = "ne"
	OpGreater        Operator = "gt"
	OpGreaterOrEqual Operator = "gte"
	OpLess           Operator = "lt"
	OpLessOrEqual    Operator = "lte"
)

type Filter struct {
	Field    Field
	Operator Operator
	Value    interface{}
}

func (f *Filter) OptionType() OptionType {
	return OptionTypeFilter
}

func (f Field) Equal(value interface{}) *Filter {
	return &Filter{
		Field:    f,
		Operator: OpEqual,
		Value:    value,
	}
}

func (f Field) NotEqual(value interface{}) *Filter {
	return &Filter{
		Field:    f,
		Operator: OpNotEqual,
		Value:    value,
	}
}

func (f Field) GreaterThan(value interface{}) *Filter {
	return &Filter{
		Field:    f,
		Operator: OpGreater,
		Value:    value,
	}
}

func (f Field) GreaterOrEqual(value interface{}) *Filter {
	return &Filter{
		Field:    f,
		Operator: OpGreaterOrEqual,
		Value:    value,
	}
}

func (f Field) LessThan(value interface{}) *Filter {
	return &Filter{
		Field:    f,
		Operator: OpLess,
		Value:    value,
	}
}

func (f Field) LessOrEqual(value interface{}) *Filter {
	return &Filter{
		Field:    f,
		Operator: OpLessOrEqual,
		Value:    value,
	}
}

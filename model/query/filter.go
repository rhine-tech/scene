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

// Eq creates an equality filter without manually constructing Field.
func Eq(field string, value interface{}) *Filter {
	return Field(field).Equal(value)
}

// Ne creates a not-equal filter.
func Ne(field string, value interface{}) *Filter {
	return Field(field).NotEqual(value)
}

// Gt creates a greater-than filter.
func Gt(field string, value interface{}) *Filter {
	return Field(field).GreaterThan(value)
}

// Gte creates a greater-or-equal filter.
func Gte(field string, value interface{}) *Filter {
	return Field(field).GreaterOrEqual(value)
}

// Lt creates a less-than filter.
func Lt(field string, value interface{}) *Filter {
	return Field(field).LessThan(value)
}

// Lte creates a less-or-equal filter.
func Lte(field string, value interface{}) *Filter {
	return Field(field).LessOrEqual(value)
}

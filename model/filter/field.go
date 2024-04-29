package filter

type FieldFilter struct {
	Field    Field
	Operator Operator
	Value    interface{}
}

func (f *FieldFilter) FilterType() FilterType {
	return FilterTypeField
}

type FieldMapper map[Field]string

var EmptyFieldMapper = make(FieldMapper)

func (m *FieldMapper) Get(field Field) string {
	val, ok := (*m)[field]
	if !ok {
		return string(field)
	}
	return val
}

type Field string

func (f Field) Equal(value interface{}) *FieldFilter {
	return &FieldFilter{
		Field:    f,
		Operator: OpEqual,
		Value:    value,
	}
}

func (f Field) NotEqual(value interface{}) *FieldFilter {
	return &FieldFilter{
		Field:    f,
		Operator: OpNotEqual,
		Value:    value,
	}
}

func (f Field) GreaterThan(value interface{}) *FieldFilter {
	return &FieldFilter{
		Field:    f,
		Operator: OpGreater,
		Value:    value,
	}
}

func (f Field) GreaterOrEqual(value interface{}) *FieldFilter {
	return &FieldFilter{
		Field:    f,
		Operator: OpGreaterOrEqual,
		Value:    value,
	}
}

func (f Field) LessThan(value interface{}) *FieldFilter {
	return &FieldFilter{
		Field:    f,
		Operator: OpLess,
		Value:    value,
	}
}

func (f Field) LessOrEqual(value interface{}) *FieldFilter {
	return &FieldFilter{
		Field:    f,
		Operator: OpLessOrEqual,
		Value:    value,
	}
}

package query

type FieldMapper map[Field]string

var EmptyFieldMapper = make(FieldMapper)

func (m *FieldMapper) Get(field Field) string {
	val, ok := (*m)[field]
	if !ok {
		return string(field)
	}
	return val
}

func (m *FieldMapper) Map(fields []Field) []string {
	result := make([]string, 0, len(fields))
	for _, f := range fields {
		mapped := m.Get(f)
		result = append(result, mapped)
	}
	return result
}

type Field string

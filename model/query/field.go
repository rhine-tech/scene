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

type Field string

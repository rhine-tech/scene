package filter

// FilterBuilder is a function that takes a filter and returns a new filter
type FilterBuilder[T any] interface {
	BuildFilter(filters ...Filter) T
	WithFieldMapper(mapper FieldMapper) FilterBuilder[T]
}

type FilterType int

const (
	FilterTypeField FilterType = iota
	FilterTypeLogical
)

// Filter is an interface.
type Filter interface {
	FilterType() FilterType
}

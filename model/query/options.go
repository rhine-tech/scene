package query

type OptionType int

const (
	OptionTypeFilter OptionType = iota
	OptionTypeLogical
	OptionTypeSort
	OptionTypeDistinct
)

// Option is an interface.
type Option interface {
	OptionType() OptionType
}

// QueryBuilder is a function that takes a filter and returns a new filter
type QueryBuilder[T any] interface {
	Build(options ...Option) T
	WithFieldMapper(mapper FieldMapper) QueryBuilder[T]
}

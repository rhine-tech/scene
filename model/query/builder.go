package query

// Builder offers a fluent API to collect query options without manually
// constructing Field/Logical types at every call site.
type Builder struct {
	options []Option
}

// NewBuilder returns an empty builder.
func NewBuilder() *Builder {
	return &Builder{
		options: make([]Option, 0),
	}
}

// Options returns the accumulated options.
func (b *Builder) Options() []Option {
	return b.options
}

// cloneOptions copies the provided slice so nested builders do not share backing arrays.
func cloneOptions(opts []Option) []Option {
	if len(opts) == 0 {
		return nil
	}
	out := make([]Option, len(opts))
	copy(out, opts)
	return out
}

// Where appends simple filters or logical options directly.
func (b *Builder) Where(opts ...Option) *Builder {
	b.options = append(b.options, opts...)
	return b
}

// Eq adds an equality filter for the provided field.
func (b *Builder) Eq(field string, value interface{}) *Builder {
	return b.Where(Eq(field, value))
}

// Ne adds a not-equal filter.
func (b *Builder) Ne(field string, value interface{}) *Builder {
	return b.Where(Ne(field, value))
}

// Gt adds a greater-than filter.
func (b *Builder) Gt(field string, value interface{}) *Builder {
	return b.Where(Gt(field, value))
}

// Gte adds a greater-or-equal filter.
func (b *Builder) Gte(field string, value interface{}) *Builder {
	return b.Where(Gte(field, value))
}

// Lt adds a less-than filter.
func (b *Builder) Lt(field string, value interface{}) *Builder {
	return b.Where(Lt(field, value))
}

// Lte adds a less-or-equal filter.
func (b *Builder) Lte(field string, value interface{}) *Builder {
	return b.Where(Lte(field, value))
}

// AndGroup creates a logical AND group using the nested builder.
func (b *Builder) AndGroup(fn func(*Builder)) *Builder {
	nested := NewBuilder()
	fn(nested)
	b.options = append(b.options, And(cloneOptions(nested.options)...))
	return b
}

// OrGroup creates a logical OR group using the nested builder.
func (b *Builder) OrGroup(fn func(*Builder)) *Builder {
	nested := NewBuilder()
	fn(nested)
	b.options = append(b.options, Or(cloneOptions(nested.options)...))
	return b
}

// NotGroup negates the nested group.
func (b *Builder) NotGroup(fn func(*Builder)) *Builder {
	nested := NewBuilder()
	fn(nested)
	b.options = append(b.options, Not(And(cloneOptions(nested.options)...)))
	return b
}

// Asc adds an ascending sort clause.
func (b *Builder) Asc(field string) *Builder {
	return b.Where(Field(field).Ascending())
}

// Desc adds a descending sort clause.
func (b *Builder) Desc(field string) *Builder {
	return b.Where(Field(field).Descending())
}

// Distinct adds a distinct clause.
func (b *Builder) Distinct(field string) *Builder {
	return b.Where(Field(field).Distinct())
}

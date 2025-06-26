package query

type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

type Order struct {
	Field Field
	Order SortOrder
}

func (o Order) OptionType() OptionType {
	return OptionTypeSort
}

func (f Field) Ascending() *Order {
	return &Order{
		Field: f,
		Order: Ascending,
	}
}

func (f Field) Descending() *Order {
	return &Order{
		Field: f,
		Order: Descending,
	}
}

package query

type Distinct struct {
	Field Field
}

func (d Distinct) OptionType() OptionType {
	return OptionTypeDistinct
}

func (f Field) Distinct() *Distinct {
	return &Distinct{
		Field: f,
	}
}

// DistinctField helps creating a distinct option without constructing Field manually.
func DistinctField(field string) *Distinct {
	return Field(field).Distinct()
}

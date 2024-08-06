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

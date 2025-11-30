package mcp

type StructureArray[T any] struct {
	Result []T `json:"result"`
	Length int `json:"length"`
}

func WrapArray[T any](array []T) StructureArray[T] {
	return StructureArray[T]{
		Result: array,
		Length: len(array),
	}
}

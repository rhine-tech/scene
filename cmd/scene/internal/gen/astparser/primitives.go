package astparser

var primitives = []string{
	"complex64", "complex128", "float32", "float64",
	"int", "int8", "int16", "int32", "int64",
	"rune", "string",
	"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
	"bool", "byte", "any", "error",
}

var _primitiveMap map[string]bool

func init() {
	_primitiveMap = make(map[string]bool)
	for _, v := range primitives {
		_primitiveMap[v] = true
	}
}

func isPrimitiveType(val string) bool {
	a, b := _primitiveMap[val]
	return a && b
}

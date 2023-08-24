package upartial

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SubStructA struct {
	SubValue0 *string
	SubValue1 *string
}

type A struct {
	Value0 *string
	ValueX *int `upartial:"Value1,default=5"`
	Value2 *bool
	Value3 *float64 `upartial:"default=4.3"`
	Value5 *[]string
	Value8 *SubStructA
}

type SubStructB struct {
	SubValue0 string
	SubValue1 string
}

type B struct {
	Value0 string
	Value1 int
	Value2 bool
	Value3 float64
	Value5 []string
	Value9 string
	Value8 SubStructB
}

func TestUpdate(t *testing.T) {
	valA := A{
		Value0: nil,
		ValueX: nil,
		Value2: nil,
		Value3: nil,
		Value5: nil,
		Value8: nil,
	}
	valB := B{
		Value0: "test",
		Value1: 1,
		Value2: true,
		Value3: 1.1,
		Value5: []string{"test"},
		Value8: SubStructB{
			SubValue0: "s0",
			SubValue1: "s1",
		},
	}
	assert.NoError(t, UpdateStruct(&valA, &valB))
	assert.Equal(t, "test", valB.Value0)
	assert.Equal(t, 5, valB.Value1)
	assert.Equal(t, 4.3, valB.Value3)
	valA.Value0 = new(string)
	*valA.Value0 = "test2"
	assert.NoError(t, UpdateStruct(&valA, &valB))
	assert.Equal(t, "test2", valB.Value0)
	valA.ValueX = new(int)
	*valA.ValueX = 10
	assert.NoError(t, UpdateStruct(&valA, &valB))
	assert.Equal(t, 10, valB.Value1)
	valA.Value8 = &SubStructA{
		SubValue0: new(string),
	}
	*valA.Value8.SubValue0 = "test3"
	assert.NoError(t, UpdateStruct(&valA, &valB))
	assert.Equal(t, "test3", valB.Value8.SubValue0)
	assert.Equal(t, "s1", valB.Value8.SubValue1)
	assert.Equal(t, "test2", valB.Value0)
}

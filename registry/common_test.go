package registry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderedregistry_AcquireAll(t *testing.T) {
	reg := NewOrderedRegistry[int, int](indexedNaming[int]())
	for i := 0; i < 10; i++ {
		reg.Register(i)
	}
	for i, v := range reg.AcquireAll() {
		assert.Equal(t, i, v)
	}
}

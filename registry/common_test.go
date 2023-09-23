package registry

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOrderedregistry_AcquireAll(t *testing.T) {
	reg := NewOrderedRegistry[int, *int](indexedNaming[*int]())
	for i := 0; i < 10; i++ {
		a := i
		reg.Register(&a)
	}
	for i, v := range reg.AcquireAll() {
		require.Equal(t, i, *v)
	}
}

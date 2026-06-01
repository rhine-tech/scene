package meta

import (
	"testing"

	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
)

func TestDummyListReturnsModuleError(t *testing.T) {
	repo := NewDummyImpl()

	result, err := repo.List("local.test", 10, 20)
	require.ErrorIs(t, err, storage.ErrLoadingMeta)
	require.Equal(t, int64(10), result.Offset)
	require.Empty(t, result.Results)
}

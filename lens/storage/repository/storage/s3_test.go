package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
)

func TestS3Storage_GetDirectURL(t *testing.T) {
	provider, err := NewS3StorageWithPresignedURLTTL(
		"localhost:9000",
		"access-key",
		"secret-key",
		"bucket",
		"default",
		false,
		true,
		"us-east-1",
		time.Minute,
	)
	require.NoError(t, err)

	storageKey := storage.NewStorageKey(provider.ProviderName(), "dir/file.txt")
	url, err := provider.GetDirectURL(storageKey)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(url, "http://localhost:9000/bucket/dir/file.txt?"), url)
	require.Contains(t, url, "X-Amz-Signature=")
}

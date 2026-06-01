package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
)

func TestS3Storage_GetPublicURLMode(t *testing.T) {
	providerAPI, err := NewS3StorageWithPublicURLMode(
		"localhost:9000",
		"access-key",
		"secret-key",
		"bucket",
		"default",
		"/api/storage/data/s3.default",
		false,
		true,
		"us-east-1",
		false,
		time.Minute,
	)
	require.NoError(t, err)
	provider := providerAPI.(*s3Storage)

	storageKey := storage.NewStorageKey(provider.ProviderName(), "dir/file.txt")
	url, err := provider.GetPublicURL(storageKey)
	require.NoError(t, err)
	require.Equal(t, "/api/storage/data/s3.default/dir/file.txt", url)

	providerAPI, err = NewS3StorageWithPublicURLMode(
		"localhost:9000",
		"access-key",
		"secret-key",
		"bucket",
		"default",
		"/api/storage/data/s3.default",
		false,
		true,
		"us-east-1",
		true,
		time.Minute,
	)
	require.NoError(t, err)
	provider = providerAPI.(*s3Storage)

	url, err = provider.GetPublicURL(storageKey)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(url, "http://localhost:9000/bucket/dir/file.txt?"), url)
	require.Contains(t, url, "X-Amz-Signature=")
}

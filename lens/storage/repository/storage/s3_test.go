package storage

import (
	"context"
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
	url, err := provider.GetDirectURL(context.Background(), storageKey)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(url, "http://localhost:9000/bucket/dir/file.txt?"), url)
	require.Contains(t, url, "X-Amz-Signature=")
}

func TestS3Storage_PreservesCanceledContextForSDKOperations(t *testing.T) {
	provider, err := NewS3Storage(
		"localhost:9000",
		"access-key",
		"secret-key",
		"bucket",
		"default",
		false,
		true,
		"us-east-1",
	)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.ErrorIs(t, provider.HealthCheck(ctx), context.Canceled)
	_, err = provider.GetDirectURL(ctx, storage.NewStorageKey(provider.ProviderName(), "file.txt"))
	require.ErrorIs(t, err, context.Canceled)
	_, err = provider.InitMultipartStore(ctx, storage.NewStorageKey(provider.ProviderName(), "file.txt"))
	require.ErrorIs(t, err, context.Canceled)
	// No remote operation is needed for an already absent upload.
	require.NoError(t, provider.AbortMultipart(ctx, "missing"))
}

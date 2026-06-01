package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type s3IntegrationEnv struct {
	provider *s3Storage
	prefix   string
}

func requireS3Integration(t *testing.T) s3IntegrationEnv {
	t.Helper()
	if os.Getenv("SCENE_STORAGE_S3_INTEGRATION") != "1" {
		t.Skip("set SCENE_STORAGE_S3_INTEGRATION=1 to run S3-compatible integration tests")
	}

	endpoint := envOr("SCENE_STORAGE_S3_ENDPOINT", "localhost:9000")
	accessKey := os.Getenv("SCENE_STORAGE_S3_ACCESS_KEY")
	secretKey := os.Getenv("SCENE_STORAGE_S3_SECRET_KEY")
	bucket := envOr("SCENE_STORAGE_S3_BUCKET", "scene-storage-integration")
	region := envOr("SCENE_STORAGE_S3_REGION", "us-east-1")
	if accessKey == "" || secretKey == "" {
		t.Fatal("SCENE_STORAGE_S3_ACCESS_KEY and SCENE_STORAGE_S3_SECRET_KEY are required")
	}

	providerAPI, err := NewS3Storage(
		endpoint,
		accessKey,
		secretKey,
		bucket,
		"integration",
		normalizeS3Endpoint(endpoint, false)+"/"+bucket,
		false,
		true,
		region,
	)
	require.NoError(t, err)

	provider, ok := providerAPI.(*s3Storage)
	require.True(t, ok)
	ensureS3Bucket(t, provider)

	return s3IntegrationEnv{
		provider: provider,
		prefix:   fmt.Sprintf("integration/%d", time.Now().UnixNano()),
	}
}

func requireS3IntegrationDirectPublicURL(t *testing.T) s3IntegrationEnv {
	t.Helper()
	if os.Getenv("SCENE_STORAGE_S3_INTEGRATION") != "1" {
		t.Skip("set SCENE_STORAGE_S3_INTEGRATION=1 to run S3-compatible integration tests")
	}

	endpoint := envOr("SCENE_STORAGE_S3_ENDPOINT", "localhost:9000")
	accessKey := os.Getenv("SCENE_STORAGE_S3_ACCESS_KEY")
	secretKey := os.Getenv("SCENE_STORAGE_S3_SECRET_KEY")
	bucket := envOr("SCENE_STORAGE_S3_BUCKET", "scene-storage-integration")
	region := envOr("SCENE_STORAGE_S3_REGION", "us-east-1")
	if accessKey == "" || secretKey == "" {
		t.Fatal("SCENE_STORAGE_S3_ACCESS_KEY and SCENE_STORAGE_S3_SECRET_KEY are required")
	}

	providerAPI, err := NewS3StorageWithPublicURLMode(
		endpoint,
		accessKey,
		secretKey,
		bucket,
		"integration",
		normalizeS3Endpoint(endpoint, false)+"/"+bucket,
		false,
		true,
		region,
		true,
		time.Minute,
	)
	require.NoError(t, err)

	provider, ok := providerAPI.(*s3Storage)
	require.True(t, ok)
	ensureS3Bucket(t, provider)

	return s3IntegrationEnv{
		provider: provider,
		prefix:   fmt.Sprintf("integration/%d", time.Now().UnixNano()),
	}
}

func envOr(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}

func ensureS3Bucket(t *testing.T, provider *s3Storage) {
	t.Helper()
	ctx := context.Background()
	_, err := provider.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(provider.bucket),
	})
	if err == nil {
		return
	}
	if !isS3NotFound(err) {
		t.Fatalf("head bucket %q: %v", provider.bucket, err)
	}
	_, err = provider.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(provider.bucket),
	})
	require.NoError(t, err)
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		_, headErr := provider.client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(provider.bucket),
		})
		require.NoError(c, headErr)
	}, 5*time.Second, 100*time.Millisecond)
}

func TestS3StorageIntegration_StoreLoadRangeMetaDelete(t *testing.T) {
	env := requireS3Integration(t)
	provider := env.provider
	key := storage.NewStorageKey(provider.ProviderName(), env.prefix, "small.txt")
	data := []byte("hello rustfs compatible s3 storage")
	t.Cleanup(func() { _ = provider.Delete(key) })

	require.NoError(t, provider.HealthCheck())
	require.NoError(t, provider.Store(key, bytes.NewReader(data)))

	meta, err := provider.Meta(key)
	require.NoError(t, err)
	require.Equal(t, key, meta.StorageKey)
	require.Equal(t, provider.ProviderName(), meta.Provider)
	require.Equal(t, key.FileID(), meta.Identifier)
	require.Equal(t, int64(len(data)), meta.ContentLength)
	require.True(t, meta.Finished)

	reader, err := provider.LoadAll(key)
	require.NoError(t, err)
	loaded, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, data, loaded)

	rangeReader, err := provider.Load(key, 6, 6)
	require.NoError(t, err)
	ranged, err := io.ReadAll(rangeReader)
	require.NoError(t, err)
	require.NoError(t, rangeReader.Close())
	require.Equal(t, []byte("rustfs"), ranged)

	_, err = provider.Load(key, -1, 1)
	require.ErrorIs(t, err, storage.ErrInvalidOffset)
	_, err = provider.Load(key, 0, 0)
	require.ErrorIs(t, err, storage.ErrInvalidLength)

	publicURL, err := provider.GetPublicURL(key)
	require.NoError(t, err)
	require.Contains(t, publicURL, key.FileID())

	require.NoError(t, provider.Delete(key))
	_, err = provider.LoadAll(key)
	require.ErrorIs(t, err, storage.ErrFileNotFound)
}

func TestS3StorageIntegration_PresignedPublicURL(t *testing.T) {
	env := requireS3IntegrationDirectPublicURL(t)
	provider := env.provider
	key := storage.NewStorageKey(provider.ProviderName(), env.prefix, "presigned.txt")
	data := []byte("presigned rustfs download")
	t.Cleanup(func() { _ = provider.Delete(key) })

	require.NoError(t, provider.Store(key, bytes.NewReader(data)))
	publicURL, err := provider.GetPublicURL(key)
	require.NoError(t, err)
	require.Contains(t, publicURL, "X-Amz-Signature=")

	resp, err := http.Get(publicURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	loaded, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, data, loaded)
}

func TestS3StorageIntegration_EmptyOverwriteNestedAndMissingObjects(t *testing.T) {
	env := requireS3Integration(t)
	provider := env.provider
	key := storage.NewStorageKey(provider.ProviderName(), env.prefix, "nested", "dir", "object.bin")
	missingKey := storage.NewStorageKey(provider.ProviderName(), env.prefix, "missing.bin")
	t.Cleanup(func() { _ = provider.Delete(key) })

	require.NoError(t, provider.Store(key, bytes.NewReader(nil)))
	meta, err := provider.Meta(key)
	require.NoError(t, err)
	require.Equal(t, int64(0), meta.ContentLength)

	reader, err := provider.LoadAll(key)
	require.NoError(t, err)
	loaded, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Empty(t, loaded)

	first := []byte("first version")
	second := []byte("second version replaces the first")
	require.NoError(t, provider.Store(key, bytes.NewReader(first)))
	require.NoError(t, provider.Store(key, bytes.NewReader(second)))

	reader, err = provider.LoadAll(key)
	require.NoError(t, err)
	loaded, err = io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, second, loaded)

	_, err = provider.Meta(missingKey)
	require.ErrorIs(t, err, storage.ErrFileNotFound)
	_, err = provider.LoadAll(missingKey)
	require.ErrorIs(t, err, storage.ErrFileNotFound)
	_, err = provider.Load(missingKey, 0, 1)
	require.ErrorIs(t, err, storage.ErrFileNotFound)

	_, err = provider.Load(key, int64(len(second)+1), 1)
	require.ErrorIs(t, err, storage.ErrInvalidOffset)
	require.NoError(t, provider.Delete(missingKey))
}

func TestS3StorageIntegration_MultipartLargeObject(t *testing.T) {
	env := requireS3Integration(t)
	provider := env.provider
	key := storage.NewStorageKey(provider.ProviderName(), env.prefix, "large.bin")
	t.Cleanup(func() { _ = provider.Delete(key) })

	uploadID, err := provider.InitMultipartStore(key)
	require.NoError(t, err)
	require.NotEmpty(t, uploadID)

	part1 := deterministicBytes(5*1024*1024, 11)
	part2 := deterministicBytes(5*1024*1024, 23)
	part3 := deterministicBytes(1024*1024+321, 37)

	// Store parts out of order to verify completion sorts part numbers before submitting.
	require.NoError(t, provider.StorePart(uploadID, 2, bytes.NewReader(part2)))
	require.NoError(t, provider.StorePart(uploadID, 1, bytes.NewReader(part1)))
	require.NoError(t, provider.StorePart(uploadID, 3, bytes.NewReader(part3)))
	require.NoError(t, provider.CompleteMultipartStore(uploadID))

	provider.uploadsLock.RLock()
	_, ok := provider.uploads[uploadID]
	provider.uploadsLock.RUnlock()
	require.False(t, ok)

	meta, err := provider.Meta(key)
	require.NoError(t, err)
	require.Equal(t, int64(len(part1)+len(part2)+len(part3)), meta.ContentLength)

	reader, err := provider.LoadAll(key)
	require.NoError(t, err)
	loaded, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	expected := make([]byte, 0, len(part1)+len(part2)+len(part3))
	expected = append(expected, part1...)
	expected = append(expected, part2...)
	expected = append(expected, part3...)
	require.Equal(t, expected, loaded)

	rangeReader, err := provider.Load(key, int64(len(part1)-16), 64)
	require.NoError(t, err)
	ranged, err := io.ReadAll(rangeReader)
	require.NoError(t, err)
	require.NoError(t, rangeReader.Close())
	expectedRange := append(part1[len(part1)-16:], part2[:48]...)
	require.Equal(t, expectedRange, ranged)
}

func TestS3StorageIntegration_MultipartAbortAndErrors(t *testing.T) {
	env := requireS3Integration(t)
	provider := env.provider
	key := storage.NewStorageKey(provider.ProviderName(), env.prefix, "aborted.bin")
	t.Cleanup(func() { _ = provider.Delete(key) })

	uploadID, err := provider.InitMultipartStore(key)
	require.NoError(t, err)
	require.NotEmpty(t, uploadID)

	require.ErrorIs(t, provider.StorePart(uploadID, 0, bytes.NewReader([]byte("bad"))), storage.ErrStorePartFailed)
	require.NoError(t, provider.StorePart(uploadID, 1, bytes.NewReader(deterministicBytes(5*1024*1024, 41))))
	require.NoError(t, provider.AbortMultipartStore(uploadID))
	require.NoError(t, provider.AbortMultipartStore(uploadID))

	require.ErrorIs(t, provider.StorePart("missing-upload", 1, bytes.NewReader([]byte("x"))), storage.ErrUploadSessionNotFound)
	err = provider.CompleteMultipartStore(uploadID)
	require.ErrorIs(t, err, storage.ErrUploadSessionNotFound)
	require.ErrorIs(t, provider.CompleteMultipartStore("missing-upload"), storage.ErrUploadSessionNotFound)

	_, err = provider.LoadAll(key)
	require.True(t, errors.Is(err, storage.ErrFileNotFound) || errors.Is(err, storage.ErrStorageFailed))
}

func TestS3StorageIntegration_MultipartEmptyAndDuplicatePart(t *testing.T) {
	env := requireS3Integration(t)
	provider := env.provider

	emptyKey := storage.NewStorageKey(provider.ProviderName(), env.prefix, "multipart-empty.bin")
	emptyUploadID, err := provider.InitMultipartStore(emptyKey)
	require.NoError(t, err)
	require.ErrorIs(t, provider.CompleteMultipartStore(emptyUploadID), storage.ErrStorePartFailed)
	require.NoError(t, provider.AbortMultipartStore(emptyUploadID))

	key := storage.NewStorageKey(provider.ProviderName(), env.prefix, "multipart-duplicate.bin")
	t.Cleanup(func() { _ = provider.Delete(key) })
	uploadID, err := provider.InitMultipartStore(key)
	require.NoError(t, err)
	require.NotEmpty(t, uploadID)

	replacedPart := deterministicBytes(5*1024*1024, 51)
	finalPart1 := deterministicBytes(5*1024*1024, 67)
	finalPart2 := deterministicBytes(1024*1024, 83)
	require.NoError(t, provider.StorePart(uploadID, 1, bytes.NewReader(replacedPart)))
	require.NoError(t, provider.StorePart(uploadID, 1, bytes.NewReader(finalPart1)))
	require.NoError(t, provider.StorePart(uploadID, 2, bytes.NewReader(finalPart2)))
	require.NoError(t, provider.CompleteMultipartStore(uploadID))

	reader, err := provider.LoadAll(key)
	require.NoError(t, err)
	loaded, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	expected := make([]byte, 0, len(finalPart1)+len(finalPart2))
	expected = append(expected, finalPart1...)
	expected = append(expected, finalPart2...)
	require.Equal(t, expected, loaded)
}

func TestS3StorageIntegration_ConcurrentSmallObjects(t *testing.T) {
	env := requireS3Integration(t)
	provider := env.provider

	const workers = 8
	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := storage.NewStorageKey(provider.ProviderName(), env.prefix, fmt.Sprintf("concurrent-%02d.bin", i))
			data := deterministicBytes(64*1024+i*17, byte(i+1))
			defer func() { _ = provider.Delete(key) }()

			if err := provider.Store(key, bytes.NewReader(data)); err != nil {
				errCh <- err
				return
			}
			reader, err := provider.LoadAll(key)
			if err != nil {
				errCh <- err
				return
			}
			loaded, err := io.ReadAll(reader)
			closeErr := reader.Close()
			if err != nil {
				errCh <- err
				return
			}
			if closeErr != nil {
				errCh <- closeErr
				return
			}
			if !bytes.Equal(data, loaded) {
				errCh <- fmt.Errorf("loaded data mismatch for object %d", i)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func deterministicBytes(size int, seed byte) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte((int(seed) + i*31 + i/251) % 256)
	}
	return data
}

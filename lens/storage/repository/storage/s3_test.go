package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
)

type capturedS3Request struct {
	path       string
	partNumber string
	uploadID   string
	body       []byte
	err        error
}

func newHTTPTestS3Storage(t *testing.T, tempDir string) (*s3Storage, <-chan capturedS3Request) {
	t.Helper()

	requests := make(chan capturedS3Request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		requests <- capturedS3Request{
			path:       r.URL.Path,
			partNumber: r.URL.Query().Get("partNumber"),
			uploadID:   r.URL.Query().Get("uploadId"),
			body:       body,
			err:        err,
		}
		w.Header().Set("ETag", `"etag"`)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	providerAPI, err := NewS3Storage(
		server.URL,
		"access-key",
		"secret-key",
		"bucket",
		"test",
		false,
		true,
		"us-east-1",
		tempDir,
	)
	require.NoError(t, err)
	provider, ok := providerAPI.(*s3Storage)
	require.True(t, ok)
	return provider, requests
}

func requireEmptyDirectory(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Empty(t, entries)
}

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
		"",
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
		"",
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

func TestS3Storage_StoreAcceptsNonSeekableReader(t *testing.T) {
	tempDir := t.TempDir()
	provider, requests := newHTTPTestS3Storage(t, tempDir)

	payload := []byte("non-seekable object")
	hash := md5.New()
	body := io.TeeReader(bytes.NewReader(payload), hash)
	key := storage.NewStorageKey(provider.ProviderName(), "object.txt")

	require.NoError(t, provider.Store(context.Background(), key, body))

	request := <-requests
	require.NoError(t, request.err)
	require.Equal(t, "/bucket/object.txt", request.path)
	require.Equal(t, payload, request.body)
	sum := md5.Sum(payload)
	require.Equal(t, sum[:], hash.Sum(nil))
	requireEmptyDirectory(t, tempDir)
}

func TestS3Storage_StoreMultipartAcceptsNonSeekableReader(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("TMPDIR", tempDir)
	provider, requests := newHTTPTestS3Storage(t, "")
	provider.uploads["upload-id"] = &s3UploadSession{
		objectKey: "multipart.txt",
		parts:     make(map[int32]types.CompletedPart),
	}

	payload := []byte("non-seekable part")
	require.NoError(t, provider.StoreMultipart(
		context.Background(),
		"upload-id",
		1,
		bytes.NewBuffer(payload),
	))

	request := <-requests
	require.NoError(t, request.err)
	require.Equal(t, "/bucket/multipart.txt", request.path)
	require.Equal(t, "1", request.partNumber)
	require.Equal(t, "upload-id", request.uploadID)
	require.Equal(t, payload, request.body)
	require.Equal(t, "etag", *provider.uploads["upload-id"].parts[1].ETag)
	requireEmptyDirectory(t, tempDir)
}

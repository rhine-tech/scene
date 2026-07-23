package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type cancelingReader struct {
	cancel context.CancelFunc
	read   bool
}

func (r *cancelingReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}
	r.read = true
	n := copy(p, "replacement")
	r.cancel()
	return n, nil
}

func TestLocalStorage_HealthCheck(t *testing.T) {
	ctx := context.Background()
	storageApi := NewLocalStorage("default", "./")
	require.NoError(t, storageApi.HealthCheck(ctx))
	storageApi = NewLocalStorage("default", "./definitely-does-not-exist")
	require.Error(t, storageApi.HealthCheck(ctx))
}

func TestLocalStorage_GetDirectURLUnsupported(t *testing.T) {
	storageApi := NewLocalStorage("default", t.TempDir())

	_, err := storageApi.GetDirectURL(context.Background(), storage.NewStorageKey(storageApi.ProviderName(), "file.txt"))
	require.ErrorIs(t, err, storage.ErrDirectURLUnsupported)
}

func TestLocalStorage_Simple(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, os.MkdirAll("./data", 0755))
	storageApi := NewLocalStorage("default", "./data")
	data := []byte("hello world")
	err := storageApi.Store(ctx, "local://test", bytes.NewBuffer(data))
	require.NoError(t, err)
	err = storageApi.Store(ctx, "local://aaa/test", bytes.NewBuffer(data))
	require.NoError(t, err)
	data2, err := storageApi.LoadAll(ctx, "local://test")
	require.NoError(t, err)
	var readed = make([]byte, len(data))
	read, err := data2.Read(readed)
	require.NoError(t, err)
	require.NoError(t, data2.Close())
	require.Equal(t, len(data), read)
	require.Equal(t, data, readed)
	data2, err = storageApi.LoadAll(ctx, "local://aaa/test")
	require.NoError(t, err)
	readed = make([]byte, len(data))
	read, err = data2.Read(readed)
	require.NoError(t, err)
	require.NoError(t, data2.Close())
	require.Equal(t, len(data), read)
	require.Equal(t, data, readed)
	require.NoError(t, storageApi.Delete(ctx, "local://test"))
	require.NoError(t, storageApi.Delete(ctx, "local://aaa/test"))
	require.NoError(t, os.RemoveAll("./data"))
}

func TestLocalStorage_StoreDoesNotOverwriteExistingFile(t *testing.T) {
	ctx := context.Background()
	basePath := t.TempDir()
	storageApi := NewLocalStorage("default", basePath)
	storageKey := storage.NewStorageKey("local.default", "safe", "object.txt")

	require.NoError(t, storageApi.Store(ctx, storageKey, bytes.NewReader([]byte("first"))))
	err := storageApi.Store(ctx, storageKey, bytes.NewReader([]byte("second")))
	require.ErrorIs(t, err, storage.ErrStorageKeyExists)

	reader, err := storageApi.LoadAll(ctx, storageKey)
	require.NoError(t, err)
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, []byte("first"), content)
}

func TestLocalStorage_Load(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, os.MkdirAll("./data", 0755))
	storageApi := NewLocalStorage("default", "./data")

	// Setup test data
	content := []byte("the quick brown fox jumps over the lazy dog")
	storageKey := storage.StorageKey("local://fox/story")
	require.NoError(t, storageApi.Store(ctx, storageKey, bytes.NewReader(content)))

	// Partial read from offset 10, length 5 ("brown")
	reader, err := storageApi.Load(ctx, storageKey, 10, 5)
	require.NoError(t, err)
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, 5, n)
	require.Equal(t, []byte("brown"), buf)

	// Offset beyond file length
	reader, err = storageApi.Load(ctx, storageKey, int64(len(content)+10), 5)
	require.ErrorIs(t, err, storage.ErrInvalidOffset)

	// negative offset
	reader, err = storageApi.Load(ctx, storageKey, -1, 5)
	require.Error(t, err)

	// Cleanup
	require.NoError(t, storageApi.Delete(ctx, storageKey))
	require.NoError(t, os.RemoveAll("./data"))
}

func TestLocalStorage_MultipartUpload(t *testing.T) {
	ctx := context.Background()
	basePath := "./data"
	require.NoError(t, os.MkdirAll(basePath, 0755))
	storageApi := NewLocalStorage("default", basePath)

	// 1. Start multipart upload
	storageKey := storage.NewStorageKey("local.default", "multi/testfile")
	uploadId, err := storageApi.InitMultipartStore(ctx, storageKey)
	require.NoError(t, err)
	require.NotEmpty(t, uploadId)

	// 2. Upload parts
	part1 := []byte("hello ")
	part2 := []byte("world!")
	require.NoError(t, storageApi.StoreMultipart(ctx, uploadId, 1, bytes.NewReader(part1)))
	require.NoError(t, storageApi.StoreMultipart(ctx, uploadId, 2, bytes.NewReader(part2)))

	// 3. Verify part files exist before completion
	part1Path := filepath.Join(basePath, fmt.Sprintf(".%s.part-%d", uploadId, 1))
	part2Path := filepath.Join(basePath, fmt.Sprintf(".%s.part-%d", uploadId, 2))
	_, err = os.Stat(part1Path)
	require.NoError(t, err)
	_, err = os.Stat(part2Path)
	require.NoError(t, err)

	// 4. Complete upload
	require.NoError(t, storageApi.CompleteMultipart(ctx, uploadId))

	// 5. Check final file exists and is concatenated correctly
	finalReader, err := storageApi.LoadAll(ctx, storageKey)
	require.NoError(t, err)
	finalContent, err := io.ReadAll(finalReader)
	require.NoError(t, err)
	require.NoError(t, finalReader.Close())
	require.Equal(t, []byte("hello world!"), finalContent)
	finalStat, err := os.Stat(filepath.Join(basePath, "multi", "testfile"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0644), finalStat.Mode().Perm())

	// 6. Check part files cleaned up (optional behavior)
	_, err = os.Stat(part1Path)
	require.True(t, os.IsNotExist(err))
	_, err = os.Stat(part2Path)
	require.True(t, os.IsNotExist(err))

	// 7. Cleanup
	require.NoError(t, storageApi.Delete(ctx, storageKey))
	require.NoError(t, os.RemoveAll(basePath))
}

func TestLocalStorage_CanceledContextStopsBeforeMutation(t *testing.T) {
	basePath := t.TempDir()
	storageAPI := NewLocalStorage("default", basePath)
	storageKey := storage.NewStorageKey(storageAPI.ProviderName(), "cancelled.txt")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.ErrorIs(t, storageAPI.HealthCheck(ctx), context.Canceled)
	require.ErrorIs(t, storageAPI.Store(ctx, storageKey, bytes.NewReader([]byte("data"))), context.Canceled)
	_, statErr := os.Stat(filepath.Join(basePath, "cancelled.txt"))
	require.True(t, os.IsNotExist(statErr))
	_, err := storageAPI.InitMultipartStore(ctx, storageKey)
	require.ErrorIs(t, err, context.Canceled)
}

func TestLocalStorage_LoadReaderObservesCancellation(t *testing.T) {
	basePath := t.TempDir()
	storageAPI := NewLocalStorage("default", basePath)
	storageKey := storage.NewStorageKey(storageAPI.ProviderName(), "stream.txt")
	require.NoError(t, storageAPI.Store(context.Background(), storageKey, bytes.NewReader([]byte("abcdef"))))

	ctx, cancel := context.WithCancel(context.Background())
	reader, err := storageAPI.LoadAll(ctx, storageKey)
	require.NoError(t, err)
	buf := make([]byte, 1)
	_, err = reader.Read(buf)
	require.NoError(t, err)
	cancel()
	_, err = reader.Read(buf)
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, reader.Close())
}

func TestLocalStorage_CanceledPartReplacementPreservesPreviousPart(t *testing.T) {
	basePath := t.TempDir()
	storageAPI := NewLocalStorage("default", basePath)
	storageKey := storage.NewStorageKey(storageAPI.ProviderName(), "multipart.bin")
	uploadID, err := storageAPI.InitMultipartStore(context.Background(), storageKey)
	require.NoError(t, err)
	require.NoError(t, storageAPI.StoreMultipart(context.Background(), uploadID, 1, bytes.NewReader([]byte("original"))))

	ctx, cancel := context.WithCancel(context.Background())
	err = storageAPI.StoreMultipart(ctx, uploadID, 1, &cancelingReader{cancel: cancel})
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, storageAPI.CompleteMultipart(context.Background(), uploadID))

	reader, err := storageAPI.LoadAll(context.Background(), storageKey)
	require.NoError(t, err)
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, []byte("original"), content)
}

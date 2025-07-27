package storage

import (
	"bytes"
	"fmt"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorage_HealthCheck(t *testing.T) {
	storageApi := NewLocalStorage("default", "./")
	require.NoError(t, storageApi.HealthCheck())
	storageApi = NewLocalStorage("default", "./definitely-does-not-exist")
	require.Error(t, storageApi.HealthCheck())
}

func TestLocalStorage_Simple(t *testing.T) {
	require.NoError(t, os.MkdirAll("./data", 0755))
	storageApi := NewLocalStorage("default", "./data")
	data := []byte("hello world")
	err := storageApi.Store("local://test", bytes.NewBuffer(data))
	require.NoError(t, err)
	err = storageApi.Store("local://aaa/test", bytes.NewBuffer(data))
	require.NoError(t, err)
	data2, err := storageApi.LoadAll("local://test")
	require.NoError(t, err)
	var readed = make([]byte, len(data))
	read, err := data2.Read(readed)
	require.NoError(t, err)
	require.Equal(t, len(data), read)
	require.Equal(t, data, readed)
	data2, err = storageApi.LoadAll("local://aaa/test")
	require.NoError(t, err)
	readed = make([]byte, len(data))
	read, err = data2.Read(readed)
	require.NoError(t, err)
	require.Equal(t, len(data), read)
	require.Equal(t, data, readed)
	require.NoError(t, storageApi.Delete("local://test"))
	require.NoError(t, storageApi.Delete("local://aaa/test"))
	require.NoError(t, os.RemoveAll("./data"))
}

func TestLocalStorage_Load(t *testing.T) {
	require.NoError(t, os.MkdirAll("./data", 0755))
	storageApi := NewLocalStorage("default", "./data")

	// Setup test data
	content := []byte("the quick brown fox jumps over the lazy dog")
	fileID := storage.FileID("local://fox/story")
	require.NoError(t, storageApi.Store(fileID, bytes.NewReader(content)))

	// Partial read from offset 10, length 5 ("brown")
	reader, err := storageApi.Load(fileID, 10, 5)
	require.NoError(t, err)
	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, []byte("brown"), buf)

	// Offset beyond file length
	reader, err = storageApi.Load(fileID, int64(len(content)+10), 5)
	require.NoError(t, err)
	buf = make([]byte, 5)
	n, err = reader.Read(buf)
	require.Equal(t, 0, n)
	require.ErrorIs(t, err, io.EOF)

	// negative offset
	reader, err = storageApi.Load(fileID, -1, 5)
	require.Error(t, err)

	// Cleanup
	require.NoError(t, storageApi.Delete(fileID))
	require.NoError(t, os.RemoveAll("./data"))
}

func TestLocalStorage_MultipartUpload(t *testing.T) {
	basePath := "./data"
	require.NoError(t, os.MkdirAll(basePath, 0755))
	storageApi := NewLocalStorage("default", basePath)

	// 1. Start multipart upload
	fileID := storage.NewFileID("local.default", "multi/testfile")
	uploadId, err := storageApi.InitMultipartStore(fileID)
	require.NoError(t, err)
	require.NotEmpty(t, uploadId)

	// 2. Upload parts
	part1 := []byte("hello ")
	part2 := []byte("world!")
	require.NoError(t, storageApi.StorePart(uploadId, 1, bytes.NewReader(part1)))
	require.NoError(t, storageApi.StorePart(uploadId, 2, bytes.NewReader(part2)))

	// 3. Verify part files exist before completion
	part1Path := filepath.Join(basePath, fmt.Sprintf(".%s.part-%d", uploadId, 1))
	part2Path := filepath.Join(basePath, fmt.Sprintf(".%s.part-%d", uploadId, 2))
	_, err = os.Stat(part1Path)
	require.NoError(t, err)
	_, err = os.Stat(part2Path)
	require.NoError(t, err)

	// 4. Complete upload
	require.NoError(t, storageApi.CompleteMultipartStore(uploadId))

	// 5. Check final file exists and is concatenated correctly
	finalReader, err := storageApi.LoadAll(fileID)
	require.NoError(t, err)
	finalContent, err := io.ReadAll(finalReader)
	require.NoError(t, err)
	require.Equal(t, []byte("hello world!"), finalContent)

	// 6. Check part files cleaned up (optional behavior)
	_, err = os.Stat(part1Path)
	require.True(t, os.IsNotExist(err))
	_, err = os.Stat(part2Path)
	require.True(t, os.IsNotExist(err))

	// 7. Cleanup
	require.NoError(t, storageApi.Delete(fileID))
	require.NoError(t, os.RemoveAll(basePath))
}

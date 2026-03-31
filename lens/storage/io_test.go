package storage

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
	"github.com/stretchr/testify/require"
)

type testLoadCall struct {
	offset int64
	length int64
}

type testStorageService struct {
	data      []byte
	loadCalls []testLoadCall
}

func (t *testStorageService) SrvImplName() scene.ImplName {
	return scene.ImplName{}
}

func (t *testStorageService) ListProviders() []string {
	return nil
}

func (t *testStorageService) ListMeta(provider string, offset, limit int64) (model.PaginationResult[FileMeta], error) {
	return model.PaginationResult[FileMeta]{}, nil
}

func (t *testStorageService) Meta(storageKey StorageKey) (FileMeta, error) {
	return FileMeta{
		StorageKey:       storageKey,
		ContentLength:    int64(len(t.data)),
		OriginalFilename: "test.bin",
	}, nil
}

func (t *testStorageService) Load(storageKey StorageKey, offset, length int64) (io.ReadCloser, error) {
	t.loadCalls = append(t.loadCalls, testLoadCall{offset: offset, length: length})
	if offset < 0 || length < 0 {
		return nil, fmt.Errorf("invalid range")
	}
	if offset > int64(len(t.data)) {
		return nil, ErrInvalidOffset
	}
	end := offset + length
	if end > int64(len(t.data)) {
		end = int64(len(t.data))
	}
	out := make([]byte, end-offset)
	copy(out, t.data[offset:end])
	return io.NopCloser(bytes.NewReader(out)), nil
}

func (t *testStorageService) LoadAll(storageKey StorageKey) (io.ReadCloser, error) {
	out := make([]byte, len(t.data))
	copy(out, t.data)
	return io.NopCloser(bytes.NewReader(out)), nil
}

func (t *testStorageService) Delete(storageKey StorageKey) error {
	return nil
}

func (t *testStorageService) Store(data io.Reader, meta FileMeta) (StorageKey, error) {
	return "", nil
}

func (t *testStorageService) StoreAt(provider, identifier string, data io.Reader, meta FileMeta) (StorageKey, error) {
	return "", nil
}

func (t *testStorageService) InitMultipartStore(provider, identifier string, meta FileMeta) (StorageKey, string, error) {
	return "", "", nil
}

func (t *testStorageService) StorePart(uploadId string, partNumber int, data io.Reader) error {
	return nil
}

func (t *testStorageService) StorePartReader(uploadId string, partNumber int, data io.Reader) error {
	return nil
}

func (t *testStorageService) CompleteMultipartStore(uploadId string) error {
	return nil
}

func (t *testStorageService) AbortMultiPartStore(uploadId string) error {
	return nil
}

func (t *testStorageService) GetPublicURL(storageKey StorageKey) (string, error) {
	return "", nil
}

func TestIoImpl_ReadAndSeekSemantics(t *testing.T) {
	svc := &testStorageService{data: []byte("abcdefghijklmnopqrstuvwxyz")}
	reader, meta, err := NewIoInterface(svc, NewStorageKey("local.test", "alphabet"))
	require.NoError(t, err)
	require.Equal(t, int64(26), meta.ContentLength)

	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, []byte("abcde"), buf)

	pos, err := reader.Seek(10, io.SeekStart)
	require.NoError(t, err)
	require.Equal(t, int64(10), pos)

	buf = make([]byte, 3)
	n, err = reader.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 3, n)
	require.Equal(t, []byte("klm"), buf)

	pos, err = reader.Seek(-3, io.SeekEnd)
	require.NoError(t, err)
	require.Equal(t, int64(23), pos)

	buf = make([]byte, 3)
	n, err = reader.Read(buf)
	require.Equal(t, 3, n)
	require.Equal(t, []byte("xyz"), buf)
	require.True(t, err == nil || err == io.EOF)

	buf = make([]byte, 1)
	n, err = reader.Read(buf)
	require.Equal(t, 0, n)
	require.ErrorIs(t, err, io.EOF)

	_, err = reader.Seek(-1, io.SeekStart)
	require.Error(t, err)
}

func TestIoImpl_ReadAheadCacheBehavior(t *testing.T) {
	data := make([]byte, defaultReadAheadSize+32)
	for i := range data {
		data[i] = byte(i % 251)
	}
	svc := &testStorageService{data: data}

	iFace, _, err := NewIoInterface(svc, NewStorageKey("local.test", "big"))
	require.NoError(t, err)

	r, ok := iFace.(*ioImpl)
	require.True(t, ok)
	_ = r

	one := make([]byte, 1)

	n, err := iFace.Read(one)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, data[0], one[0])
	require.Len(t, svc.loadCalls, 1)
	require.Equal(t, int64(0), svc.loadCalls[0].offset)
	require.Equal(t, defaultReadAheadSize, svc.loadCalls[0].length)

	n, err = iFace.Read(one)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, data[1], one[0])
	require.Len(t, svc.loadCalls, 1)

	_, err = iFace.Seek(128, io.SeekStart)
	require.NoError(t, err)
	n, err = iFace.Read(one)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, data[128], one[0])
	require.Len(t, svc.loadCalls, 1)

	_, err = iFace.Seek(defaultReadAheadSize, io.SeekStart)
	require.NoError(t, err)
	n, err = iFace.Read(one)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, data[defaultReadAheadSize], one[0])
	require.Len(t, svc.loadCalls, 2)
	require.Equal(t, defaultReadAheadSize, svc.loadCalls[1].offset)
}

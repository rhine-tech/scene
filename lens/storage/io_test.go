package storage

import (
	"bytes"
	"context"
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
	closes    int
}

type testReadCloser struct {
	reader  io.Reader
	onClose func()
	closed  bool
}

func (r *testReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *testReadCloser) Close() error {
	if !r.closed {
		r.closed = true
		r.onClose()
	}
	return nil
}

func (t *testStorageService) SrvImplName() scene.ImplName {
	return scene.ImplName{}
}

func (t *testStorageService) ListProviders() []string {
	return nil
}

func (t *testStorageService) ListMeta(context.Context, string, int64, int64) (model.PaginationResult[FileMeta], error) {
	return model.PaginationResult[FileMeta]{}, nil
}

func (t *testStorageService) Meta(_ context.Context, storageKey StorageKey) (FileMeta, error) {
	return FileMeta{
		StorageKey:       storageKey,
		ContentLength:    int64(len(t.data)),
		OriginalFilename: "test.bin",
	}, nil
}

func (t *testStorageService) Load(_ context.Context, _ StorageKey, offset, length int64) (io.ReadCloser, error) {
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
	return &testReadCloser{
		reader:  bytes.NewReader(out),
		onClose: func() { t.closes++ },
	}, nil
}

func (t *testStorageService) LoadAll(context.Context, StorageKey) (io.ReadCloser, error) {
	out := make([]byte, len(t.data))
	copy(out, t.data)
	return io.NopCloser(bytes.NewReader(out)), nil
}

func (t *testStorageService) Delete(context.Context, StorageKey) error {
	return nil
}

func (t *testStorageService) Store(context.Context, io.Reader, FileMeta) (StorageKey, error) {
	return "", nil
}

func (t *testStorageService) StoreAt(context.Context, string, string, io.Reader, FileMeta) (StorageKey, error) {
	return "", nil
}

func (t *testStorageService) InitMultipartStore(context.Context, string, string, FileMeta) (StorageKey, string, error) {
	return "", "", nil
}

func (t *testStorageService) StoreMultipart(context.Context, string, int, io.Reader) error {
	return nil
}

func (t *testStorageService) CompleteMultipart(context.Context, string) (FileMeta, error) {
	return FileMeta{}, nil
}

func (t *testStorageService) AbortMultipart(context.Context, string) error {
	return nil
}

func (t *testStorageService) GetDirectURL(context.Context, StorageKey) (string, error) {
	return "", nil
}

func TestContentReaderReadAndSeekSemantics(t *testing.T) {
	svc := &testStorageService{data: []byte("abcdefghijklmnopqrstuvwxyz")}
	reader, meta, err := OpenContent(context.Background(), svc, NewStorageKey("local.test", "alphabet"))
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
	require.NoError(t, reader.Close())
	require.NoError(t, reader.Close())
	require.Equal(t, 3, svc.closes)
	require.Equal(t, []testLoadCall{
		{offset: 0, length: 26},
		{offset: 10, length: 16},
		{offset: 23, length: 3},
	}, svc.loadCalls)
}

func TestContentReaderSequentialReadsReuseProviderReader(t *testing.T) {
	data := make([]byte, 2<<20+32)
	for i := range data {
		data[i] = byte(i % 251)
	}
	svc := &testStorageService{data: data}

	reader, _, err := OpenContent(context.Background(), svc, NewStorageKey("local.test", "big"))
	require.NoError(t, err)

	loaded, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, data, loaded)
	require.Len(t, svc.loadCalls, 1)
	require.Equal(t, testLoadCall{offset: 0, length: int64(len(data))}, svc.loadCalls[0])
	require.Zero(t, svc.closes)
	require.NoError(t, reader.Close())
	require.Equal(t, 1, svc.closes)
}

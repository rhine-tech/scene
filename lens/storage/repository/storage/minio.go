package storage

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/storage"
	"io"
	"net/url"
	"strings"
)

type minioStorage struct {
	client    *minio.Client
	bucket    string
	name      string
	urlPrefix string
}

func (m *minioStorage) InitMultipartStore(fileId storage.FileID) (uploadId string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *minioStorage) StorePart(uploadId string, partNumber int, data io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (m *minioStorage) CompleteMultipartStore(uploadId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *minioStorage) AbortMultipartStore(uploadId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *minioStorage) Meta(fileId storage.FileID) (meta storage.FileMeta, err error) {
	return meta, storage.ErrLoadingMeta
}

func NewMinioStorage(endpoint, accessKey, secretKey, bucket, name, urlPrefix string, useSSL bool) (storage.IStorageProvider, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &minioStorage{
		client:    client,
		bucket:    bucket,
		name:      name,
		urlPrefix: strings.TrimRight(urlPrefix, "/"),
	}, nil
}

func (m *minioStorage) ProviderName() string {
	return "minio." + m.name
}

func (m *minioStorage) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IStorageProvider", "minio")
}

func (m *minioStorage) HealthCheck() error {
	ctx := context.Background()
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return storage.ErrStorageError.WithDetail(err)
	}
	if !exists {
		return storage.ErrStorageError.WithDetailStr("bucket does not exist")
	}

	//// Test write permissions by uploading a small test file
	//testKey := fmt.Sprintf("healthcheck-%d.tmp", time.Now().UnixNano())
	//testContent := []byte("healthcheck")
	//
	//_, err = m.client.PutObject(ctx, m.bucket, testKey,
	//	io.NopCloser(bytes.NewReader(testContent)),
	//	int64(len(testContent)),
	//	minio.PutObjectOptions{})
	//if err != nil {
	//	return fmt.Errorf("write test failed: %w", err)
	//}
	//
	//// Clean up test file
	//err = m.client.RemoveObject(ctx, m.bucket, testKey, minio.RemoveObjectOptions{})
	//if err != nil {
	//	return fmt.Errorf("failed to clean up test file: %w", err)
	//}
	return nil
}

func (m *minioStorage) Store(fileId storage.FileID, data io.Reader) error {
	key := fileId.ID()
	_, err := m.client.PutObject(context.Background(), m.bucket, key, data, -1, minio.PutObjectOptions{})
	if err != nil {
		return storage.ErrStorageFailed.WithDetail(err)
	}
	return nil
}

func (m *minioStorage) Load(fileId storage.FileID, offset, length int64) (io.Reader, error) {
	opts := minio.GetObjectOptions{}
	if offset > 0 || length > 0 {
		err := opts.SetRange(offset, offset+length-1)
		if err != nil {
			return nil, storage.ErrStorageFailed.WithDetail(err)
		}
	} else {
		return nil, storage.ErrInvalidOffset
	}
	obj, err := m.client.GetObject(context.Background(), m.bucket, fileId.ID(), opts)
	if err != nil {
		return nil, storage.ErrFileNotFound.WithDetail(err)
	}
	return obj, nil
}

func (m *minioStorage) LoadAll(fileId storage.FileID) (io.Reader, error) {
	obj, err := m.client.GetObject(context.Background(), m.bucket, fileId.ID(), minio.GetObjectOptions{})
	if err != nil {
		return nil, storage.ErrFileNotFound.WithDetail(err)
	}
	return obj, nil
}

func (m *minioStorage) Delete(fileId storage.FileID) error {
	err := m.client.RemoveObject(context.Background(), m.bucket, fileId.ID(), minio.RemoveObjectOptions{})
	if err != nil {
		return storage.ErrStorageFailed.WithDetail(err)
	}
	return nil
}

func (m *minioStorage) GetPublicURL(fileId storage.FileID) (string, error) {
	key := fileId.ID()
	return url.JoinPath(m.urlPrefix, key)
}

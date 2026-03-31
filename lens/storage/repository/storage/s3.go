package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/storage"
)

type s3UploadSession struct {
	objectKey string
	parts     map[int32]types.CompletedPart
	partsLock sync.Mutex
}

type s3Storage struct {
	client    *s3.Client
	bucket    string
	name      string
	urlPrefix string

	uploads     map[string]*s3UploadSession
	uploadsLock sync.RWMutex
}

func NewS3Storage(
	endpoint, accessKey, secretKey, bucket, name, urlPrefix string,
	useSSL, forcePathStyle bool,
	region string,
) (storage.IStorageProvider, error) {
	if region == "" {
		region = "us-east-1"
	}
	awsCfg := aws.Config{
		Region: region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
		if endpoint != "" {
			o.BaseEndpoint = aws.String(normalizeS3Endpoint(endpoint, useSSL))
		}
	})
	if client == nil {
		return nil, errors.New("failed to initialize s3 client")
	}
	return &s3Storage{
		client:    client,
		bucket:    bucket,
		name:      name,
		urlPrefix: strings.TrimRight(urlPrefix, "/"),
		uploads:   make(map[string]*s3UploadSession),
	}, nil
}

func (s *s3Storage) ProviderName() string {
	return "s3." + s.name
}

func (s *s3Storage) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IStorageProvider", "s3")
}

func (s *s3Storage) HealthCheck() error {
	_, err := s.client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return storage.ErrStorageError.WithDetail(err)
	}
	return nil
}

func (s *s3Storage) Meta(storageKey storage.StorageKey) (storage.FileMeta, error) {
	out, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey.FileID()),
	})
	if err != nil {
		if isS3NotFound(err) {
			return storage.FileMeta{}, storage.ErrFileNotFound
		}
		return storage.FileMeta{}, storage.ErrStorageError.WithDetail(err)
	}
	return storage.FileMeta{
		StorageKey:       storageKey,
		Provider:         storageKey.Provider(),
		Identifier:       storageKey.FileID(),
		OriginalFilename: storageKey.FileID(),
		ContentType:      aws.ToString(out.ContentType),
		ContentLength:    aws.ToInt64(out.ContentLength),
		Md5Checksum:      strings.Trim(aws.ToString(out.ETag), "\""),
		Finished:         true,
		CreatedAt:        safeLastModified(out.LastModified),
		UpdatedAt:        safeLastModified(out.LastModified),
	}, nil
}

func (s *s3Storage) Store(storageKey storage.StorageKey, data io.Reader) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey.FileID()),
		Body:   data,
	})
	if err != nil {
		return storage.ErrStorageFailed.WithDetail(err)
	}
	return nil
}

func (s *s3Storage) Load(storageKey storage.StorageKey, offset, length int64) (io.ReadCloser, error) {
	if offset < 0 {
		return nil, storage.ErrInvalidOffset
	}
	if length <= 0 {
		return nil, storage.ErrInvalidLength
	}
	resp, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey.FileID()),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, storage.ErrFileNotFound
		}
		return nil, storage.ErrStorageFailed.WithDetail(err)
	}
	return resp.Body, nil
}

func (s *s3Storage) LoadAll(storageKey storage.StorageKey) (io.ReadCloser, error) {
	resp, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey.FileID()),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, storage.ErrFileNotFound
		}
		return nil, storage.ErrStorageFailed.WithDetail(err)
	}
	return resp.Body, nil
}

func (s *s3Storage) Delete(storageKey storage.StorageKey) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey.FileID()),
	})
	if err != nil {
		if isS3NotFound(err) {
			return storage.ErrFileNotFound
		}
		return storage.ErrStorageFailed.WithDetail(err)
	}
	return nil
}

func (s *s3Storage) InitMultipartStore(storageKey storage.StorageKey) (string, error) {
	resp, err := s.client.CreateMultipartUpload(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey.FileID()),
	})
	if err != nil {
		return "", storage.ErrInitPartUploadFailed.WithDetail(err)
	}
	uploadId := aws.ToString(resp.UploadId)
	s.uploadsLock.Lock()
	s.uploads[uploadId] = &s3UploadSession{
		objectKey: storageKey.FileID(),
		parts:     make(map[int32]types.CompletedPart),
	}
	s.uploadsLock.Unlock()
	return uploadId, nil
}

func (s *s3Storage) StorePart(uploadId string, partNumber int, data io.Reader) error {
	if partNumber <= 0 {
		return storage.ErrStorePartFailed.WithDetailStr("invalid part number")
	}
	s.uploadsLock.RLock()
	session, ok := s.uploads[uploadId]
	s.uploadsLock.RUnlock()
	if !ok {
		return storage.ErrUploadSessionNotFound
	}
	pn := int32(partNumber)
	resp, err := s.client.UploadPart(context.Background(), &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(session.objectKey),
		UploadId:   aws.String(uploadId),
		PartNumber: aws.Int32(pn),
		Body:       data,
	})
	if err != nil {
		return storage.ErrStorePartFailed.WithDetail(err)
	}
	session.partsLock.Lock()
	session.parts[pn] = types.CompletedPart{
		PartNumber: aws.Int32(pn),
		ETag:       aws.String(strings.Trim(aws.ToString(resp.ETag), "\"")),
	}
	session.partsLock.Unlock()
	return nil
}

func (s *s3Storage) CompleteMultipartStore(uploadId string) error {
	s.uploadsLock.RLock()
	session, ok := s.uploads[uploadId]
	s.uploadsLock.RUnlock()
	if !ok {
		return storage.ErrUploadSessionNotFound
	}
	session.partsLock.Lock()
	partNumbers := make([]int32, 0, len(session.parts))
	for n := range session.parts {
		partNumbers = append(partNumbers, n)
	}
	sort.Slice(partNumbers, func(i, j int) bool { return partNumbers[i] < partNumbers[j] })
	parts := make([]types.CompletedPart, 0, len(partNumbers))
	for _, n := range partNumbers {
		parts = append(parts, session.parts[n])
	}
	session.partsLock.Unlock()
	if len(parts) == 0 {
		return storage.ErrStorePartFailed.WithDetailStr("no uploaded parts")
	}
	_, err := s.client.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(session.objectKey),
		UploadId: aws.String(uploadId),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		return storage.ErrStorePartFailed.WithDetail(err)
	}
	s.uploadsLock.Lock()
	delete(s.uploads, uploadId)
	s.uploadsLock.Unlock()
	return nil
}

func (s *s3Storage) AbortMultipartStore(uploadId string) error {
	s.uploadsLock.RLock()
	session, ok := s.uploads[uploadId]
	s.uploadsLock.RUnlock()
	if !ok {
		return nil
	}
	_, err := s.client.AbortMultipartUpload(context.Background(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(session.objectKey),
		UploadId: aws.String(uploadId),
	})
	if err != nil && !isS3NotFound(err) {
		return storage.ErrFailToAbortMultipartStore.WithDetail(err)
	}
	s.uploadsLock.Lock()
	delete(s.uploads, uploadId)
	s.uploadsLock.Unlock()
	return nil
}

func (s *s3Storage) GetPublicURL(storageKey storage.StorageKey) (string, error) {
	return url.JoinPath(s.urlPrefix, storageKey.FileID())
}

func isS3NotFound(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.ErrorCode() {
	case "NotFound", "NoSuchKey", "NoSuchUpload", "NoSuchBucket", "404":
		return true
	default:
		return false
	}
}

func safeLastModified(v *time.Time) time.Time {
	if v == nil {
		return time.Time{}
	}
	return *v
}

func normalizeS3Endpoint(endpoint string, useSSL bool) string {
	ep := strings.TrimSpace(endpoint)
	if ep == "" {
		return ep
	}
	if strings.HasPrefix(ep, "http://") || strings.HasPrefix(ep, "https://") {
		return ep
	}
	if useSSL {
		return "https://" + ep
	}
	return "http://" + ep
}

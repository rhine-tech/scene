package storage

import (
	"context"
	"io"
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
)

const Lens scene.ModuleName = "storage"

type FileMeta struct {
	StorageKey       StorageKey `json:"storage_key"`
	Provider         string     `json:"provider"`
	Identifier       string     `json:"identifier"`
	OriginalFilename string     `json:"original_filename"`
	ContentType      string     `json:"content_type"`
	ContentLength    int64      `json:"content_length"`
	Md5Checksum      string     `json:"md5_checksum"`
	Finished         bool       `json:"finished"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (f *FileMeta) FillMissing() FileMeta {
	if !IsStorageKey(string(f.StorageKey)) {
		return *f
	}
	if f.Provider == "" {
		f.Provider = f.StorageKey.Provider()
	}
	if f.Identifier == "" {
		f.Identifier = f.StorageKey.FileID()
	}
	return *f
}

type IStorageProvider interface {
	scene.Named
	// ProviderName return provider name in storageKey
	ProviderName() string
	// HealthCheck will check if this Storage provider is accessible
	HealthCheck(ctx context.Context) error
	// Meta will return any possible metadata can be found, as a fallback option if IFileMetaRepository failed
	Meta(ctx context.Context, storageKey StorageKey) (meta FileMeta, err error)
	Load(ctx context.Context, storageKey StorageKey, offset, length int64) (reader io.ReadCloser, err error)
	LoadAll(ctx context.Context, storageKey StorageKey) (reader io.ReadCloser, err error)
	// Store will store the data in the storage at path and return the storageKey,
	// if path not exists, it will create the path.
	Store(ctx context.Context, storageKey StorageKey, data io.Reader) (err error)
	Delete(ctx context.Context, storageKey StorageKey) error
	// Multipart related
	InitMultipartStore(ctx context.Context, storageKey StorageKey) (uploadId string, err error)
	StoreMultipart(ctx context.Context, uploadId string, partNumber int, data io.Reader) error
	CompleteMultipart(ctx context.Context, uploadId string) error
	AbortMultipart(ctx context.Context, uploadId string) error
	// GetDirectURL returns a URL that accesses the provider directly.
	// Providers that cannot expose direct URLs return ErrDirectURLUnsupported.
	GetDirectURL(ctx context.Context, storageKey StorageKey) (url string, err error)
}

type IFileMetaRepository interface {
	scene.Named
	// Store will store the metadata in the repository
	// will overwrite the old metadata if exists
	Store(ctx context.Context, meta FileMeta) error
	Load(ctx context.Context, storageKey StorageKey) (meta FileMeta, err error)
	Delete(ctx context.Context, storageKey StorageKey) error
	List(ctx context.Context, provider string, offset, limit int64) (model.PaginationResult[FileMeta], error)
}

// IStorageService is the storage module's service boundary.
// Every non-nil error returned directly by a service method must be a storage errcode;
// provider and repository errors, including context errors, are mapped at this boundary.
type IStorageService interface {
	scene.Service
	ListProviders() []string
	// ListMeta will list meta from specific provider.
	ListMeta(ctx context.Context, provider string, offset, limit int64) (model.PaginationResult[FileMeta], error)
	// Meta return the meta given a storageKey
	Meta(ctx context.Context, storageKey StorageKey) (meta FileMeta, err error)
	// Load will load file stream at offset with length.
	// Caller must close the returned reader.
	Load(ctx context.Context, storageKey StorageKey, offset, length int64) (reader io.ReadCloser, err error)
	// LoadAll will load full file stream.
	// Caller must close the returned reader.
	LoadAll(ctx context.Context, storageKey StorageKey) (reader io.ReadCloser, err error)
	Delete(ctx context.Context, storageKey StorageKey) error
	// Store will store data at default provider
	// it calls StoreAt internally
	// Store consumes data from reader until EOF.
	Store(ctx context.Context, data io.Reader, meta FileMeta) (storageKey StorageKey, err error)
	// StoreAt will store data using the given provider and identifier.
	// If provider is empty, the service default provider will be used.
	// If identifier is empty, the service will generate one internally.
	StoreAt(ctx context.Context, provider, identifier string, data io.Reader, meta FileMeta) (storageKey StorageKey, err error)
	// Multipart related
	// InitMultipartStore will initialize a multipart upload using the given provider and identifier.
	// If provider is empty, the service default provider will be used.
	// If identifier is empty, the service will generate one internally.
	// It returns both the resolved storage key and the upload id.
	InitMultipartStore(ctx context.Context, provider, identifier string, meta FileMeta) (storageKey StorageKey, uploadId string, err error)
	StoreMultipart(ctx context.Context, uploadId string, partNumber int, data io.Reader) error
	// CompleteMultipart completes an upload and returns its finalized metadata.
	CompleteMultipart(ctx context.Context, uploadId string) (FileMeta, error)
	AbortMultipart(ctx context.Context, uploadId string) error
	// GetDirectURL returns a URL that accesses the storage provider directly.
	GetDirectURL(ctx context.Context, storageKey StorageKey) (url string, err error)
	// todo ListMultipartParts(uploadId)
}

// UploadSession contains info to resume/complete uploads
type UploadSession struct {
	StorageKey StorageKey `json:"storage_key"`
	Created    time.Time  `json:"created"`
}

type IUploadSessionTracker interface {
	scene.Named
	Save(ctx context.Context, uploadId string, session UploadSession) error
	Get(ctx context.Context, uploadId string) (UploadSession, error)
	Delete(ctx context.Context, uploadId string) error
}

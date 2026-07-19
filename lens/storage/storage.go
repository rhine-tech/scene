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
	HealthCheck() error
	// Meta will return any possible metadata can be found, as a fallback option if IFileMetaRepository failed
	Meta(storageKey StorageKey) (meta FileMeta, err error)
	Load(storageKey StorageKey, offset, length int64) (reader io.ReadCloser, err error)
	LoadAll(storageKey StorageKey) (reader io.ReadCloser, err error)
	// Store will store the data in the storage at path and return the storageKey,
	// if path not exists, it will create the path.
	Store(storageKey StorageKey, data io.Reader) (err error)
	Delete(storageKey StorageKey) error
	// Multipart related
	InitMultipartStore(storageKey StorageKey) (uploadId string, err error)
	StorePart(uploadId string, partNumber int, data io.Reader) error
	CompleteMultipartStore(uploadId string) error
	AbortMultipartStore(uploadId string) error
	// GetPublicURL get public url which can be access in public network
	GetPublicURL(storageKey StorageKey) (url string, err error)
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

type IStorageService interface {
	scene.Service
	ListProviders() []string
	// ListMeta will list meta from specific provider.
	ListMeta(provider string, offset, limit int64) (model.PaginationResult[FileMeta], error)
	// Meta return the meta given a storageKey
	Meta(storageKey StorageKey) (meta FileMeta, err error)
	// Load will load file stream at offset with length.
	// Caller must close the returned reader.
	Load(storageKey StorageKey, offset, length int64) (reader io.ReadCloser, err error)
	// LoadAll will load full file stream.
	// Caller must close the returned reader.
	LoadAll(storageKey StorageKey) (reader io.ReadCloser, err error)
	Delete(storageKey StorageKey) error
	// Store will store data at default provider
	// it calls StoreAt internally
	// Store consumes data from reader until EOF.
	Store(data io.Reader, meta FileMeta) (storageKey StorageKey, err error)
	// StoreAt will store data using the given provider and identifier.
	// If provider is empty, the service default provider will be used.
	// If identifier is empty, the service will generate one internally.
	StoreAt(provider, identifier string, data io.Reader, meta FileMeta) (storageKey StorageKey, err error)
	// Multipart related
	// InitMultipartStore will initialize a multipart upload using the given provider and identifier.
	// If provider is empty, the service default provider will be used.
	// If identifier is empty, the service will generate one internally.
	// It returns both the resolved storage key and the upload id.
	InitMultipartStore(provider, identifier string, meta FileMeta) (storageKey StorageKey, uploadId string, err error)
	StorePart(uploadId string, partNumber int, data io.Reader) error
	StorePartReader(uploadId string, partNumber int, data io.Reader) error
	CompleteMultipartStore(uploadId string) error
	AbortMultiPartStore(uploadId string) error
	// GetPublicURL get public url which can be access in public network
	GetPublicURL(storageKey StorageKey) (url string, err error)
	// todo ListMultipartParts(uploadId)
}

// UploadSession contains info to resume/complete uploads
type UploadSession struct {
	StorageKey StorageKey `json:"storage_key"`
	Created    time.Time  `json:"created"`
}

type IUploadSessionTracker interface {
	scene.Named
	Save(uploadId string, session UploadSession) error
	Get(uploadId string) (UploadSession, error)
	Delete(uploadId string) error
}

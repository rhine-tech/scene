package storage

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/model"
	"io"
	"strings"
	"time"
)

const Lens scene.ModuleName = "storage"

// FileID is the unique identifier of a file in storage.
// it is composed with {Provider}://{ID}
// example: tos.buketName://objectName
// example: local.name://objectName
type FileID string

func NewFileID(provider string, path ...string) FileID {
	return FileID(provider + "://" + strings.TrimPrefix(strings.Join(path, "/"), "/"))
}

func ParseFileID(fileId string) (FileID, bool) {
	parts := strings.Split(fileId, "://")
	if len(parts) != 2 {
		return "", false
	}
	return FileID(fileId), true
}

func IsFileID(fileId string) bool {
	parts := strings.Split(fileId, "://")
	return len(parts) == 2
}

func (f FileID) Provider() string {
	return strings.Split(string(f), "://")[0]
}

func (f FileID) ID() string {
	val := strings.Split(string(f), "://")
	if len(val) != 2 {
		return ""
	}
	return val[1]
}

type FileMeta struct {
	FileID           FileID    `gorm:"primaryKey;column:file_id" json:"file_id"`
	Provider         string    `gorm:"column:provider" json:"provider"`
	Identifier       string    `gorm:"column:identifier" json:"identifier"`
	OriginalFilename string    `json:"original_filename" gorm:"column:original_filename"`
	ContentType      string    `json:"content_type" gorm:"column:content_type"`
	ContentLength    int64     `json:"content_length" gorm:"column:content_length"`
	Md5Checksum      string    `json:"md5_checksum" gorm:"column:md5_checksum"`
	Finished         bool      `json:"finished" gorm:"column:finished"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (f *FileMeta) FillMissing() FileMeta {
	if !IsFileID(string(f.FileID)) {
		return *f
	}
	if f.Provider == "" {
		f.Provider = f.FileID.Provider()
	}
	if f.Identifier == "" {
		f.Identifier = f.FileID.ID()
	}
	return *f
}

type IStorageProvider interface {
	scene.Named
	// ProviderName return provider name in fileId
	ProviderName() string
	// HealthCheck will check if this Storage provider is accessible
	HealthCheck() error
	// Meta will return any possible metadata can be found, as a fallback option if IFileMetaRepository failed
	Meta(fileId FileID) (meta FileMeta, err error)
	Load(fileId FileID, offset, length int64) (reader io.Reader, err error)
	LoadAll(fileId FileID) (reader io.Reader, err error)
	// Store will store the data in the storage at path and return the fileId,
	// if path not exists, it will create the path.
	Store(fileId FileID, data io.Reader) (err error)
	Delete(fileId FileID) error
	// Multipart related
	InitMultipartStore(fileId FileID) (uploadId string, err error)
	StorePart(uploadId string, partNumber int, data io.Reader) error
	CompleteMultipartStore(uploadId string) error
	AbortMultipartStore(uploadId string) error
	// GetPublicURL get public url which can be access in public network
	GetPublicURL(fileId FileID) (url string, err error)
}

type IFileMetaRepository interface {
	scene.Named
	// Store will store the metadata in the repository
	// will overwrite the old metadata if exists
	Store(meta FileMeta) error
	Load(fileId FileID) (meta FileMeta, err error)
	Delete(fileId FileID) error
	List(provider string, offset, limit int64) (model.PaginationResult[FileMeta], error)
}

type IStorageService interface {
	scene.Service
	ListProviders() []string
	// ListMeta will list meta from specific provider.
	ListMeta(provider string, offset, limit int64) (model.PaginationResult[FileMeta], error)
	// Meta return the meta given a fileId
	Meta(fileId FileID) (meta FileMeta, err error)
	// Load will load file data at offset with length.
	// basically io.Seeker & io.Reader
	Load(fileId FileID, offset, length int64) (data []byte, err error)
	// LoadAll will load data, if reach end of file it will return io.EOF
	// only use when you know the size of the data
	LoadAll(fileId FileID) (data []byte, err error)
	Delete(fileId FileID) error
	// Store will store data at default provider
	// it calls StoreAt internally
	Store(data []byte, meta FileMeta) (fileId FileID, err error)
	// StoreAt will store the data in the storage at path using specified provider,
	// if provider is empty, it will use default provider
	StoreAt(provider string, data []byte, meta FileMeta) (fileId FileID, err error)
	// Multipart related
	InitMultipartStore(fileId FileID, meta FileMeta) (uploadId string, err error)
	StorePart(uploadId string, partNumber int, data []byte) error
	StorePartReader(uploadId string, partNumber int, data io.Reader) error
	CompleteMultipartStore(uploadId string) error
	AbortMultiPartStore(uploadId string) error
	// todo ListMultipartParts(uploadId)
}

// UploadSession contains info to resume/complete uploads
type UploadSession struct {
	FileID  FileID    `json:"file_id"`
	Created time.Time `json:"created"`
}

type IUploadSessionTracker interface {
	scene.Named
	Save(uploadId string, session UploadSession) error
	Get(uploadId string) (UploadSession, error)
	Delete(uploadId string) error
}

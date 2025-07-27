package service

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/errcode"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
	"io"
	"sync"
	"time"
)

// StorageService implements IStorageService.
type StorageService struct {
	defaultProvider string
	providers       map[string]storage.IStorageProvider
	metaRepo        storage.IFileMetaRepository   `aperture:""`
	uploadSessions  storage.IUploadSessionTracker `aperture:""`
	log             logger.ILogger                `aperture:""`
	pvdrLock        sync.Mutex
	pvdrNames       []string
}

func (s *StorageService) Setup() error {
	s.log.Info("storage service setup")
	return nil
}

func (s *StorageService) SrvImplName() scene.ImplName {
	return storage.Lens.ImplNameNoVer("IStorageService")
}

// NewStorageService creates a new StorageService with a default provider.
func NewStorageService(
	metaRepo storage.IFileMetaRepository,
	sessionRepo storage.IUploadSessionTracker,
	defaultProvider string, providers ...storage.IStorageProvider) *StorageService {
	providersMap := make(map[string]storage.IStorageProvider)
	for _, p := range providers {
		providersMap[p.ProviderName()] = p
	}
	if _, exists := providersMap[defaultProvider]; !exists {
		panic("default provider not found")
	}
	return &StorageService{
		defaultProvider: defaultProvider,
		providers:       providersMap,
		metaRepo:        metaRepo,
		uploadSessions:  sessionRepo,
	}
}

// ListProviders returns the names of the available providers.
func (s *StorageService) ListProviders() []string {
	if s.pvdrNames != nil {
		return s.pvdrNames
	}
	s.pvdrLock.Lock()
	var providerNames []string
	for name := range s.providers {
		providerNames = append(providerNames, name)
	}
	s.pvdrNames = providerNames
	s.pvdrLock.Unlock()
	return s.pvdrNames
}

// Store stores data using the default provider.
func (s *StorageService) Store(data []byte, meta storage.FileMeta) (fileId storage.FileID, err error) {
	return s.StoreAt(s.defaultProvider, data, meta)
}

// StoreAt stores data at a specific path and provider.
func (s *StorageService) StoreAt(provider string, data []byte, meta storage.FileMeta) (fileId storage.FileID, err error) {
	if provider == "" {
		provider = s.defaultProvider
	}
	storageProvider, exists := s.providers[provider]
	if !exists {
		return "", storage.ErrStorageNotFound
	}
	md5Sum := md5.Sum(data)
	fileKey := randString(4) + hex.EncodeToString(md5Sum[:])
	fileId = storage.NewFileID(provider, fileKey)
	err = storageProvider.Store(fileId, io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		s.log.ErrorW("failed to store file", "fileId", fileId, "err", err)
		return "", storage.ErrFailToStore
	}
	meta.Finished = true
	meta.FillMissing()
	meta.Md5Checksum = hex.EncodeToString(md5Sum[:])
	err = s.metaRepo.Store(meta)
	if err != nil {
		return "", err
	}
	s.log.InfoW("file stored", "fileId", fileId)
	return fileId, nil
}

// Meta retrieves the metadata of a file based on fileId.
func (s *StorageService) Meta(fileId storage.FileID) (meta storage.FileMeta, err error) {
	meta, err = s.metaRepo.Load(fileId)
	if err != nil {
		s.log.ErrorW("failed to load meta, fallback to provider meta", "fileId", fileId, "err", err)
	}
	storager, exists := s.providers[fileId.Provider()]
	if !exists {
		return meta, storage.ErrLoadingMeta.WrapIfNot(err)
	}
	meta, err2 := storager.Meta(fileId)
	if err2 != nil {
		s.log.ErrorW("fail to load meta from storage provider", "fileId", fileId, "err", err2)
		return meta, storage.ErrLoadingMeta.WrapIfNot(err)
	}
	meta.FillMissing()
	return meta, nil
}

// Load retrieves data based on fileId.
func (s *StorageService) Load(fileId storage.FileID, offset, length int64) ([]byte, error) {
	storager, exists := s.providers[fileId.Provider()]
	if !exists {
		return nil, storage.ErrStorageNotFound
	}

	reader, err := storager.Load(fileId, offset, length)
	if err != nil {
		s.log.ErrorW("failed to load file", "fileId", fileId, "err", err)
		return nil, errcode.Must(err, storage.ErrFailToLoad)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		s.log.ErrorW("failed to read data", "fileId", fileId, "err", err)
		return nil, storage.ErrFailToLoad
	}

	return data, nil
}

// LoadAll retrieves data based on fileId.
func (s *StorageService) LoadAll(fileId storage.FileID) ([]byte, error) {
	storager, exists := s.providers[fileId.Provider()]
	if !exists {
		return nil, storage.ErrStorageNotFound
	}

	reader, err := storager.LoadAll(fileId)
	if err != nil {
		s.log.ErrorW("failed to load file", "fileId", fileId, "err", err)
		return nil, errcode.Must(err, storage.ErrFailToLoad)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		s.log.ErrorW("failed to read data", "fileId", fileId, "err", err)
		return nil, storage.ErrFailToLoad
	}

	return data, nil
}

// Delete deletes a file based on fileId.
func (s *StorageService) Delete(fileId storage.FileID) error {
	storager, exists := s.providers[fileId.Provider()]
	if !exists {
		return storage.ErrStorageNotFound
	}
	err := storager.Delete(fileId)
	if err != nil {
		s.log.ErrorW("failed to delete file", "fileId", fileId, "err", err)
		return storage.ErrFailToDelete
	}
	err = s.metaRepo.Delete(fileId)
	if err != nil {
		s.log.ErrorW("failed to delete file meta", "fileId", fileId, "err", err)
	}
	return nil
}

func (s *StorageService) InitMultipartStore(fileId storage.FileID, meta storage.FileMeta) (string, error) {
	pvd, ok := s.providers[fileId.Provider()]
	if !ok {
		return "", storage.ErrStorageNotFound
	}
	uploadId, err := pvd.InitMultipartStore(fileId)
	if err != nil {
		s.log.ErrorW("failed to initiate multipart upload", "fileId", fileId, "err", err)
		return "", storage.ErrInitPartUploadFailed.WrapIfNot(err)
	}
	err = s.uploadSessions.Save(uploadId, storage.UploadSession{
		FileID:  fileId,
		Created: time.Now(),
	})
	if err != nil {
		err2 := pvd.AbortMultipartStore(uploadId)
		if err2 != nil {
			s.log.ErrorW("failed to abort multipart upload", "fileId", fileId, "err", err2)
		}
		return "", err
	}
	meta.Finished = false
	meta.FillMissing()
	err = s.metaRepo.Store(meta)
	if err != nil {
		s.log.ErrorW("failed to store multipart upload", "fileId", fileId, "err", err)
		// cancel store
		err2 := pvd.AbortMultipartStore(uploadId)
		if err2 != nil {
			s.log.ErrorW("failed to abort multipart upload", "fileId", fileId, "err", err2)
		}
		return "", storage.ErrInitPartUploadFailed.WrapIfNot(err)
	}
	return uploadId, nil
}

func (s *StorageService) StorePart(uploadId string, partNumber int, data []byte) error {
	return s.StorePartReader(uploadId, partNumber, bytes.NewReader(data))
}

func (s *StorageService) StorePartReader(uploadId string, partNumber int, data io.Reader) error {
	get, err := s.uploadSessions.Get(uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		return storage.ErrUploadSessionNotFound
	}
	pvd, ok := s.providers[get.FileID.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}
	err = pvd.StorePart(uploadId, partNumber, data)
	if err != nil {
		s.log.ErrorW("failed to store part upload", "uploadId", uploadId, "err", err)
		return storage.ErrStorePartFailed.WrapIfNot(err)
	}
	return nil
}

func (s *StorageService) CompleteMultipartStore(uploadId string) error {
	get, err := s.uploadSessions.Get(uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		return storage.ErrUploadSessionNotFound
	}

	pvd, ok := s.providers[get.FileID.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}

	err = pvd.CompleteMultipartStore(uploadId)
	if err != nil {
		s.log.ErrorW("failed to complete multipart upload", "uploadId", uploadId, "err", err)
		return storage.ErrStorePartFailed.WrapIfNot(err)
	}

	// update meta
	meta, err := s.metaRepo.Load(get.FileID)
	if err != nil {
		s.log.ErrorW("failed to load meta", "fileId", get.FileID, "err", err)
	} else {
		meta.Finished = true
		err = s.metaRepo.Store(meta)
		if err != nil {
			s.log.ErrorW("failed to store meta for multipart upload", "fileId", get.FileID, "err", err)
		}
	}

	if err := s.uploadSessions.Delete(uploadId); err != nil {
		s.log.WarnW("failed to cleanup upload session after complete", "uploadId", uploadId, "err", err)
	}

	return nil
}

func (s *StorageService) AbortMultiPartStore(uploadId string) error {
	get, err := s.uploadSessions.Get(uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		return storage.ErrUploadSessionNotFound
	}

	pvd, ok := s.providers[get.FileID.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}

	err = pvd.AbortMultipartStore(uploadId)
	if err != nil {
		s.log.ErrorW("failed to abort multipart upload", "uploadId", uploadId, "err", err)
		return storage.ErrStorePartFailed.WrapIfNot(err)
	}

	if err := s.uploadSessions.Delete(uploadId); err != nil {
		s.log.WarnW("failed to cleanup upload session after abort", "uploadId", uploadId, "err", err)
	}

	return nil
}

func (s *StorageService) ListMeta(provider string, offset, limit int64) (model.PaginationResult[storage.FileMeta], error) {
	_, ok := s.providers[provider]
	if !ok {
		return model.PaginationResult[storage.FileMeta]{}, storage.ErrStorageNotFound
	}
	reuslt, err := s.metaRepo.List(provider, offset, limit)
	return reuslt, storage.ErrFailToListMeta.WrapIfNot(err)
}

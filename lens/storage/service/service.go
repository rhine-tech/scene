package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/model"
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
func (s *StorageService) Store(data io.Reader, meta storage.FileMeta) (storageKey storage.StorageKey, err error) {
	return s.StoreAt("", "", data, meta)
}

// StoreAt stores data at a specific path and provider.
func (s *StorageService) StoreAt(provider, identifier string, data io.Reader, meta storage.FileMeta) (storageKey storage.StorageKey, err error) {
	storageKey, err = s.resolveStorageKey(provider, identifier)
	if err != nil {
		return "", err
	}
	storageProvider, exists := s.providers[storageKey.Provider()]
	if !exists {
		return "", storage.ErrStorageNotFound
	}
	hash := md5.New()
	reader := io.TeeReader(data, hash)
	err = storageProvider.Store(storageKey, reader)
	if err != nil {
		s.log.ErrorW("failed to store file", "storageKey", storageKey, "err", err)
		if errors.Is(err, storage.ErrStorageKeyExists) {
			return "", storage.ErrStorageKeyExists
		}
		return "", storage.ErrFailToStore
	}
	meta.Finished = true
	meta.StorageKey = storageKey
	meta.Provider = storageKey.Provider()
	meta.Identifier = storageKey.FileID()
	meta.FillMissing()
	meta.Md5Checksum = hex.EncodeToString(hash.Sum(nil))
	err = s.metaRepo.Store(meta)
	if err != nil {
		s.log.ErrorW("failed to store file meta", "storageKey", storageKey, "err", err)
		return "", storage.ErrFailToStore
	}
	s.log.InfoW("file stored", "storageKey", storageKey)
	return storageKey, nil
}

// Meta retrieves the metadata of a file based on storageKey.
func (s *StorageService) Meta(storageKey storage.StorageKey) (meta storage.FileMeta, err error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return meta, err
	}
	meta, err = s.metaRepo.Load(storageKey)
	if err == nil {
		meta.FillMissing()
		return meta, nil
	}
	s.log.ErrorW("failed to load meta, fallback to provider meta", "storageKey", storageKey, "err", err)
	meta, err2 := storager.Meta(storageKey)
	if err2 != nil {
		s.log.ErrorW("fail to load meta from storage provider", "storageKey", storageKey, "err", err2)
		if errors.Is(err2, storage.ErrFileNotFound) {
			return meta, storage.ErrFileNotFound
		}
		return meta, storage.ErrLoadingMeta
	}
	meta.FillMissing()
	return meta, nil
}

// Load retrieves data based on storageKey.
func (s *StorageService) Load(storageKey storage.StorageKey, offset, length int64) (io.ReadCloser, error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return nil, err
	}

	reader, err := storager.Load(storageKey, offset, length)
	if err != nil {
		s.log.ErrorW("failed to load file", "storageKey", storageKey, "err", err)
		return nil, loadError(err)
	}
	return reader, nil
}

// LoadAll retrieves data based on storageKey.
func (s *StorageService) LoadAll(storageKey storage.StorageKey) (io.ReadCloser, error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return nil, err
	}

	reader, err := storager.LoadAll(storageKey)
	if err != nil {
		s.log.ErrorW("failed to load file", "storageKey", storageKey, "err", err)
		return nil, loadError(err)
	}
	return reader, nil
}

// Delete deletes a file based on storageKey.
func (s *StorageService) Delete(storageKey storage.StorageKey) error {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return err
	}
	err = storager.Delete(storageKey)
	if err != nil {
		s.log.ErrorW("failed to delete file", "storageKey", storageKey, "err", err)
		if errors.Is(err, storage.ErrFileNotFound) {
			return storage.ErrFileNotFound
		}
		return storage.ErrFailToDelete
	}
	err = s.metaRepo.Delete(storageKey)
	if err != nil {
		s.log.ErrorW("failed to delete file meta", "storageKey", storageKey, "err", err)
	}
	return nil
}

func (s *StorageService) InitMultipartStore(provider, identifier string, meta storage.FileMeta) (storage.StorageKey, string, error) {
	storageKey, err := s.resolveStorageKey(provider, identifier)
	if err != nil {
		return "", "", err
	}
	pvd, ok := s.providers[storageKey.Provider()]
	if !ok {
		return "", "", storage.ErrStorageNotFound
	}
	_, err = s.metaRepo.Load(storageKey)
	if err == nil {
		return "", "", storage.ErrStorageKeyExists
	}
	if !errors.Is(err, storage.ErrMetaNotFound) {
		s.log.ErrorW("failed to check multipart upload meta", "storageKey", storageKey, "err", err)
		return "", "", storage.ErrInitPartUploadFailed
	}
	uploadId, err := pvd.InitMultipartStore(storageKey)
	if err != nil {
		s.log.ErrorW("failed to initiate multipart upload", "storageKey", storageKey, "err", err)
		return "", "", storage.ErrInitPartUploadFailed
	}
	err = s.uploadSessions.Save(uploadId, storage.UploadSession{
		StorageKey: storageKey,
		Created:    time.Now(),
	})
	if err != nil {
		err2 := pvd.AbortMultipartStore(uploadId)
		if err2 != nil {
			s.log.ErrorW("failed to abort multipart upload", "storageKey", storageKey, "err", err2)
		}
		s.log.ErrorW("failed to save multipart upload session", "storageKey", storageKey, "uploadId", uploadId, "err", err)
		return "", "", storage.ErrInitPartUploadFailed
	}
	meta.Finished = false
	meta.StorageKey = storageKey
	meta.Provider = storageKey.Provider()
	meta.Identifier = storageKey.FileID()
	meta.FillMissing()
	err = s.metaRepo.Store(meta)
	if err != nil {
		s.log.ErrorW("failed to store multipart upload", "storageKey", storageKey, "err", err)
		// cancel store
		err2 := pvd.AbortMultipartStore(uploadId)
		if err2 != nil {
			s.log.ErrorW("failed to abort multipart upload", "storageKey", storageKey, "err", err2)
		}
		return "", "", storage.ErrInitPartUploadFailed
	}
	return storageKey, uploadId, nil
}

func (s *StorageService) StorePart(uploadId string, partNumber int, data io.Reader) error {
	return s.StorePartReader(uploadId, partNumber, data)
}

func (s *StorageService) StorePartReader(uploadId string, partNumber int, data io.Reader) error {
	get, err := s.uploadSessions.Get(uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		return storage.ErrUploadSessionNotFound
	}
	pvd, ok := s.providers[get.StorageKey.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}
	err = pvd.StorePart(uploadId, partNumber, data)
	if err != nil {
		s.log.ErrorW("failed to store part upload", "uploadId", uploadId, "err", err)
		if errors.Is(err, storage.ErrUploadSessionNotFound) {
			return storage.ErrUploadSessionNotFound
		}
		return storage.ErrStorePartFailed
	}
	return nil
}

func (s *StorageService) CompleteMultipartStore(uploadId string) error {
	get, err := s.uploadSessions.Get(uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		return storage.ErrUploadSessionNotFound
	}

	pvd, ok := s.providers[get.StorageKey.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}

	err = pvd.CompleteMultipartStore(uploadId)
	if err != nil {
		s.log.ErrorW("failed to complete multipart upload", "uploadId", uploadId, "err", err)
		if errors.Is(err, storage.ErrUploadSessionNotFound) {
			return storage.ErrUploadSessionNotFound
		}
		return storage.ErrStorePartFailed
	}

	// update meta
	meta, err := s.metaRepo.Load(get.StorageKey)
	if err != nil {
		s.log.ErrorW("failed to load meta", "storageKey", get.StorageKey, "err", err)
	} else {
		meta.Finished = true
		err = s.metaRepo.Store(meta)
		if err != nil {
			s.log.ErrorW("failed to store meta for multipart upload", "storageKey", get.StorageKey, "err", err)
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

	pvd, ok := s.providers[get.StorageKey.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}

	err = pvd.AbortMultipartStore(uploadId)
	if err != nil {
		s.log.ErrorW("failed to abort multipart upload", "uploadId", uploadId, "err", err)
		return storage.ErrFailToAbortMultipartStore
	}

	if err := s.uploadSessions.Delete(uploadId); err != nil {
		s.log.WarnW("failed to cleanup upload session after abort", "uploadId", uploadId, "err", err)
	}

	return nil
}

func (s *StorageService) resolveStorageKey(provider, identifier string) (storage.StorageKey, error) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = s.defaultProvider
	}
	identifier = storage.NormalizeIdentifier(identifier)
	if identifier == "" {
		identifier = randString(4) + strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	storageKey := storage.NewStorageKey(provider, identifier)
	if err := storage.ValidateStorageKey(storageKey); err != nil {
		return "", err
	}
	if _, exists := s.providers[provider]; !exists {
		return "", storage.ErrStorageNotFound
	}
	return storageKey, nil
}

func (s *StorageService) ListMeta(provider string, offset, limit int64) (model.PaginationResult[storage.FileMeta], error) {
	_, ok := s.providers[provider]
	if !ok {
		return model.PaginationResult[storage.FileMeta]{}, storage.ErrStorageNotFound
	}
	reuslt, err := s.metaRepo.List(provider, offset, limit)
	if err != nil {
		s.log.ErrorW("failed to list file meta", "provider", provider, "offset", offset, "limit", limit, "err", err)
		return model.PaginationResult[storage.FileMeta]{}, storage.ErrFailToListMeta
	}
	return reuslt, nil
}

func (s *StorageService) GetPublicURL(storageKey storage.StorageKey) (string, error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return "", err
	}
	url, err := storager.GetPublicURL(storageKey)
	if err != nil {
		s.log.ErrorW("failed to get public url", "storageKey", storageKey, "err", err)
		return "", storage.ErrStorageError
	}
	return url, nil
}

func (s *StorageService) providerFor(storageKey storage.StorageKey) (storage.IStorageProvider, error) {
	if err := storage.ValidateStorageKey(storageKey); err != nil {
		return nil, err
	}
	storager, exists := s.providers[storageKey.Provider()]
	if !exists {
		return nil, storage.ErrStorageNotFound
	}
	return storager, nil
}

func loadError(err error) error {
	switch {
	case errors.Is(err, storage.ErrFileNotFound):
		return storage.ErrFileNotFound
	case errors.Is(err, storage.ErrInvalidStorageKey):
		return storage.ErrInvalidStorageKey
	case errors.Is(err, storage.ErrInvalidOffset):
		return storage.ErrInvalidOffset
	case errors.Is(err, storage.ErrInvalidLength):
		return storage.ErrInvalidLength
	default:
		return storage.ErrFailToLoad
	}
}

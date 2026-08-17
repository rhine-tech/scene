package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"slices"
	"strings"
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
	providerNames   []string
}

const finalizationTimeout = 5 * time.Second

func (s *StorageService) Setup() error {
	s.log.Info("storage service setup")
	return nil
}

func (s *StorageService) TearDown() error {
	return nil
}

func (s *StorageService) ImplName() scene.ImplName {
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
	providerNames := make([]string, 0, len(providersMap))
	for name := range providersMap {
		providerNames = append(providerNames, name)
	}
	slices.Sort(providerNames)
	return &StorageService{
		defaultProvider: defaultProvider,
		providers:       providersMap,
		providerNames:   providerNames,
		metaRepo:        metaRepo,
		uploadSessions:  sessionRepo,
	}
}

// ListProviders returns the names of the available providers.
func (s *StorageService) ListProviders() []string {
	return slices.Clone(s.providerNames)
}

// Store stores data using the default provider.
func (s *StorageService) Store(ctx context.Context, data io.Reader, meta storage.FileMeta) (storageKey storage.StorageKey, err error) {
	return s.StoreAt(ctx, "", "", data, meta)
}

// StoreAt stores data at a specific path and provider.
func (s *StorageService) StoreAt(ctx context.Context, provider, identifier string, data io.Reader, meta storage.FileMeta) (storageKey storage.StorageKey, err error) {
	storageKey, storageProvider, err := s.resolveStorageKey(provider, identifier)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	reader := io.TeeReader(data, hash)
	err = storageProvider.Store(ctx, storageKey, reader)
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
	finalizeCtx, cancel := finalizationContext(ctx)
	defer cancel()
	err = s.metaRepo.Store(finalizeCtx, meta)
	if err != nil {
		s.log.ErrorW("failed to store file meta", "storageKey", storageKey, "err", err)
		return "", storage.ErrFailToStore
	}
	s.log.InfoW("file stored", "storageKey", storageKey)
	return storageKey, nil
}

// Meta retrieves the metadata of a file based on storageKey.
func (s *StorageService) Meta(ctx context.Context, storageKey storage.StorageKey) (meta storage.FileMeta, err error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return meta, err
	}
	meta, err = s.metaRepo.Load(ctx, storageKey)
	if err == nil {
		meta.FillMissing()
		return meta, nil
	}
	if isContextError(err) {
		return meta, storage.ErrLoadingMeta
	}
	s.log.ErrorW("failed to load meta, fallback to provider meta", "storageKey", storageKey, "err", err)
	meta, err2 := storager.Meta(ctx, storageKey)
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
func (s *StorageService) Load(ctx context.Context, storageKey storage.StorageKey, offset, length int64) (io.ReadCloser, error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return nil, err
	}

	reader, err := storager.Load(ctx, storageKey, offset, length)
	if err != nil {
		s.log.ErrorW("failed to load file", "storageKey", storageKey, "err", err)
		return nil, loadError(err)
	}
	return reader, nil
}

// LoadAll retrieves data based on storageKey.
func (s *StorageService) LoadAll(ctx context.Context, storageKey storage.StorageKey) (io.ReadCloser, error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return nil, err
	}

	reader, err := storager.LoadAll(ctx, storageKey)
	if err != nil {
		s.log.ErrorW("failed to load file", "storageKey", storageKey, "err", err)
		return nil, loadError(err)
	}
	return reader, nil
}

// Delete deletes a file based on storageKey.
func (s *StorageService) Delete(ctx context.Context, storageKey storage.StorageKey) error {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return err
	}
	err = storager.Delete(ctx, storageKey)
	if err != nil {
		s.log.ErrorW("failed to delete file", "storageKey", storageKey, "err", err)
		if errors.Is(err, storage.ErrFileNotFound) {
			return storage.ErrFileNotFound
		}
		return storage.ErrFailToDelete
	}
	finalizeCtx, cancel := finalizationContext(ctx)
	defer cancel()
	err = s.metaRepo.Delete(finalizeCtx, storageKey)
	if err != nil {
		s.log.ErrorW("failed to delete file meta", "storageKey", storageKey, "err", err)
		return storage.ErrFailToDelete
	}
	return nil
}

func (s *StorageService) InitMultipartStore(ctx context.Context, provider, identifier string, meta storage.FileMeta) (storage.StorageKey, string, error) {
	storageKey, pvd, err := s.resolveStorageKey(provider, identifier)
	if err != nil {
		return "", "", err
	}
	_, err = s.metaRepo.Load(ctx, storageKey)
	if err == nil {
		return "", "", storage.ErrStorageKeyExists
	}
	if !errors.Is(err, storage.ErrMetaNotFound) {
		s.log.ErrorW("failed to check multipart upload meta", "storageKey", storageKey, "err", err)
		return "", "", storage.ErrInitPartUploadFailed
	}
	uploadId, err := pvd.InitMultipartStore(ctx, storageKey)
	if err != nil {
		s.log.ErrorW("failed to initiate multipart upload", "storageKey", storageKey, "err", err)
		return "", "", storage.ErrInitPartUploadFailed
	}
	err = s.uploadSessions.Save(ctx, uploadId, storage.UploadSession{
		StorageKey: storageKey,
		Created:    time.Now(),
	})
	if err != nil {
		cleanupCtx, cancel := finalizationContext(ctx)
		defer cancel()
		err2 := pvd.AbortMultipart(cleanupCtx, uploadId)
		if err2 != nil {
			s.log.ErrorW("failed to abort multipart upload", "storageKey", storageKey, "err", err2)
		} else if err2 := s.uploadSessions.Delete(cleanupCtx, uploadId); err2 != nil {
			s.log.WarnW("failed to cleanup multipart upload session", "uploadId", uploadId, "err", err2)
		}
		s.log.ErrorW("failed to save multipart upload session", "storageKey", storageKey, "uploadId", uploadId, "err", err)
		return "", "", storage.ErrInitPartUploadFailed
	}
	meta.Finished = false
	meta.StorageKey = storageKey
	meta.Provider = storageKey.Provider()
	meta.Identifier = storageKey.FileID()
	meta.FillMissing()
	err = s.metaRepo.Store(ctx, meta)
	if err != nil {
		s.log.ErrorW("failed to store multipart upload", "storageKey", storageKey, "err", err)
		cleanupCtx, cancel := finalizationContext(ctx)
		defer cancel()
		err2 := pvd.AbortMultipart(cleanupCtx, uploadId)
		if err2 != nil {
			s.log.ErrorW("failed to abort multipart upload", "storageKey", storageKey, "err", err2)
		} else if err2 := s.uploadSessions.Delete(cleanupCtx, uploadId); err2 != nil {
			s.log.WarnW("failed to cleanup multipart upload session", "uploadId", uploadId, "err", err2)
		}
		return "", "", storage.ErrInitPartUploadFailed
	}
	return storageKey, uploadId, nil
}

func (s *StorageService) StoreMultipart(ctx context.Context, uploadId string, partNumber int, data io.Reader) error {
	get, err := s.uploadSessions.Get(ctx, uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		if isContextError(err) {
			return storage.ErrStorePartFailed
		}
		return storage.ErrUploadSessionNotFound
	}
	pvd, ok := s.providers[get.StorageKey.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}
	err = pvd.StoreMultipart(ctx, uploadId, partNumber, data)
	if err != nil {
		s.log.ErrorW("failed to store part upload", "uploadId", uploadId, "err", err)
		if errors.Is(err, storage.ErrUploadSessionNotFound) {
			return storage.ErrUploadSessionNotFound
		}
		return storage.ErrStorePartFailed
	}
	return nil
}

func (s *StorageService) CompleteMultipart(ctx context.Context, uploadId string) (storage.FileMeta, error) {
	get, err := s.uploadSessions.Get(ctx, uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		if isContextError(err) {
			return storage.FileMeta{}, storage.ErrStorePartFailed
		}
		return storage.FileMeta{}, storage.ErrUploadSessionNotFound
	}

	pvd, ok := s.providers[get.StorageKey.Provider()]
	if !ok {
		return storage.FileMeta{}, storage.ErrStorageNotFound
	}

	err = pvd.CompleteMultipart(ctx, uploadId)
	if err != nil {
		s.log.ErrorW("failed to complete multipart upload", "uploadId", uploadId, "err", err)
		if errors.Is(err, storage.ErrUploadSessionNotFound) {
			return storage.FileMeta{}, storage.ErrUploadSessionNotFound
		}
		return storage.FileMeta{}, storage.ErrStorePartFailed
	}

	finalizeCtx, cancel := finalizationContext(ctx)
	defer cancel()

	// The object is already committed remotely; finish bookkeeping even if the request was canceled.
	meta, err := s.metaRepo.Load(finalizeCtx, get.StorageKey)
	if err != nil {
		s.log.ErrorW("failed to load meta", "storageKey", get.StorageKey, "err", err)
		return storage.FileMeta{}, storage.ErrStorePartFailed
	}
	meta.Finished = true
	err = s.metaRepo.Store(finalizeCtx, meta)
	if err != nil {
		s.log.ErrorW("failed to store meta for multipart upload", "storageKey", get.StorageKey, "err", err)
		return storage.FileMeta{}, storage.ErrStorePartFailed
	}

	if err := s.uploadSessions.Delete(finalizeCtx, uploadId); err != nil {
		s.log.WarnW("failed to cleanup upload session after complete", "uploadId", uploadId, "err", err)
	}

	return meta, nil
}

func (s *StorageService) AbortMultipart(ctx context.Context, uploadId string) error {
	get, err := s.uploadSessions.Get(ctx, uploadId)
	if err != nil {
		s.log.ErrorW("failed to get upload session", "uploadId", uploadId, "err", err)
		if isContextError(err) {
			return storage.ErrFailToAbortMultipartStore
		}
		return storage.ErrUploadSessionNotFound
	}

	pvd, ok := s.providers[get.StorageKey.Provider()]
	if !ok {
		return storage.ErrStorageNotFound
	}

	err = pvd.AbortMultipart(ctx, uploadId)
	if err != nil {
		s.log.ErrorW("failed to abort multipart upload", "uploadId", uploadId, "err", err)
		return storage.ErrFailToAbortMultipartStore
	}

	finalizeCtx, cancel := finalizationContext(ctx)
	defer cancel()
	if err := s.uploadSessions.Delete(finalizeCtx, uploadId); err != nil {
		s.log.WarnW("failed to cleanup upload session after abort", "uploadId", uploadId, "err", err)
		return storage.ErrFailToAbortMultipartStore
	}

	return nil
}

func (s *StorageService) resolveStorageKey(provider, identifier string) (storage.StorageKey, storage.IStorageProvider, error) {
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
		return "", nil, err
	}
	pvd, exists := s.providers[provider]
	if !exists {
		return "", nil, storage.ErrStorageNotFound
	}
	return storageKey, pvd, nil
}

func (s *StorageService) ListMeta(ctx context.Context, provider string, offset, limit int64) (model.PaginationResult[storage.FileMeta], error) {
	_, ok := s.providers[provider]
	if !ok {
		return model.PaginationResult[storage.FileMeta]{}, storage.ErrStorageNotFound
	}
	reuslt, err := s.metaRepo.List(ctx, provider, offset, limit)
	if err != nil {
		s.log.ErrorW("failed to list file meta", "provider", provider, "offset", offset, "limit", limit, "err", err)
		return model.PaginationResult[storage.FileMeta]{}, storage.ErrFailToListMeta
	}
	return reuslt, nil
}

func (s *StorageService) GetDirectURL(ctx context.Context, storageKey storage.StorageKey) (string, error) {
	storager, err := s.providerFor(storageKey)
	if err != nil {
		return "", err
	}
	url, err := storager.GetDirectURL(ctx, storageKey)
	if err != nil {
		s.log.ErrorW("failed to get direct URL", "storageKey", storageKey, "err", err)
		if errors.Is(err, storage.ErrDirectURLUnsupported) {
			return "", storage.ErrDirectURLUnsupported
		}
		return "", storage.ErrGetDirectURLFailed
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

func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func finalizationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), finalizationTimeout)
}

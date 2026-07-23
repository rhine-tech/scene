package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/storage"
)

type uploadSession struct {
	storageKey storage.StorageKey
	uploadId   string
	parts      map[int]string // partNumber -> file path
	createdAt  time.Time
	partsMutex sync.Mutex
}

type localStorage struct {
	name        string
	localPath   string
	log         logger.ILogger `aperture:""`
	uploads     map[string]*uploadSession
	uploadsLock sync.RWMutex
}

type sectionReadCloser struct {
	ctx    context.Context
	reader io.Reader
	closer io.Closer
}

func (s *sectionReadCloser) Read(p []byte) (int, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	return s.reader.Read(p)
}

func (s *sectionReadCloser) Close() error {
	return s.closer.Close()
}

func (l *localStorage) ProviderName() string {
	return "local." + l.name
}

func (l *localStorage) HealthCheck(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stat, err := os.Stat(l.localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return storage.ErrStorageError.WithDetailStr("local path does not exist")
		}
		return storage.ErrStorageError.WithDetail(err)
	}
	if !stat.IsDir() {
		return storage.ErrStorageError.WithDetailStr("path is not a directory")
	}
	// Check read permission.
	file, err := os.Open(l.localPath)
	if err == nil {
		_ = file.Close()
	} else if os.IsPermission(err) {
		return storage.ErrStorageError.WithDetailStr("no read permission")
	} else {
		return storage.ErrStorageError.WithDetail(err)
	}

	// Check write and delete permission using a unique temporary file.
	tempFile, err := os.CreateTemp(l.localPath, ".healthcheck-*.tmp")
	if err != nil {
		return storage.ErrStorageError.WithDetailStr("no write permission")
	}
	tempFilename := tempFile.Name()
	if _, err := tempFile.Write([]byte("test")); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFilename)
		return storage.ErrStorageError.WithDetailStr("no write permission")
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempFilename)
		return storage.ErrStorageError.WithDetailStr("no write permission")
	}
	if err := os.Remove(tempFilename); err != nil {
		return storage.ErrStorageError.WithDetailStr("no delete permission")
	}

	return nil
}

func (l *localStorage) Setup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	l.log.Infof("local storage init with path: %s error: %v", l.localPath, l.HealthCheck(ctx))
	return nil
}

func NewLocalStorage(name string, localPath string) storage.IStorageProvider {
	return &localStorage{
		name:      name,
		localPath: localPath,
		uploads:   make(map[string]*uploadSession),
	}
}

func (l *localStorage) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IStorageProvider", "local")
}

func (l *localStorage) Meta(ctx context.Context, storageKey storage.StorageKey) (storage.FileMeta, error) {
	if err := ctx.Err(); err != nil {
		return storage.FileMeta{}, err
	}
	pathParts := l.cleanupPath(storageKey)
	if len(pathParts) == 0 {
		return storage.FileMeta{}, storage.ErrInvalidStorageKey
	}

	path := filepath.Join(append([]string{l.localPath}, pathParts...)...)
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return storage.FileMeta{}, storage.ErrFileNotFound
		}
		return storage.FileMeta{}, storage.ErrStorageError.WithDetail(err)
	}
	// Try to detect MIME type using first N bytes
	f, err := os.Open(path)
	if err != nil {
		return storage.FileMeta{}, storage.ErrStorageError.WithDetail(err)
	}

	header := make([]byte, 512)
	n, _ := f.Read(header)
	_ = f.Close()

	meta := storage.FileMeta{
		StorageKey:       storageKey,
		OriginalFilename: filepath.Base(path),
		ContentType:      http.DetectContentType(header[:n]),
		ContentLength:    stat.Size(),
		Md5Checksum:      "",
		Finished:         true,
		CreatedAt:        stat.ModTime(),
		UpdatedAt:        stat.ModTime(),
	}

	return meta, nil
}

func (l *localStorage) Store(ctx context.Context, storageKey storage.StorageKey, data io.Reader) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	prefixs := l.cleanupPath(storageKey)
	if len(prefixs) == 0 {
		return storage.ErrStorageFailed.WithDetailStr("invalid_prefix")
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return storage.ErrStorageFailed.WithDetail(err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return storage.ErrStorageKeyExists
		}
		return storage.ErrStorageFailed.WithDetail(err)
	}
	_, err = copyWithContext(ctx, file, data)
	if err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		if err := operationContextError(err); err != nil {
			return err
		}
		return storage.ErrStorageFailed.WithDetail(err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return storage.ErrStorageFailed.WithDetail(err)
	}
	return nil
}

func (l *localStorage) Load(ctx context.Context, storageKey storage.StorageKey, offset, length int64) (reader io.ReadCloser, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	prefixs := l.cleanupPath(storageKey)
	if len(prefixs) == 0 {
		return nil, storage.ErrStorageFailed
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)

	if offset < 0 {
		return nil, storage.ErrInvalidOffset
	}

	// local ranged load requires a positive length
	if length <= 0 {
		return nil, storage.ErrInvalidLength
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrFileNotFound
		}
		return nil, storage.ErrStorageError.WithDetail(err)
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, storage.ErrStorageError.WithDetailStr("failed to stat file")
	}
	if offset > stat.Size() {
		_ = file.Close()
		return nil, storage.ErrInvalidOffset
	}
	maxLen := stat.Size() - offset
	if length > maxLen {
		length = maxLen
	}

	section := io.NewSectionReader(file, offset, length)
	return &sectionReadCloser{
		ctx:    ctx,
		reader: section,
		closer: file,
	}, nil
}

func (l *localStorage) LoadAll(ctx context.Context, storageKey storage.StorageKey) (reader io.ReadCloser, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	prefixs := l.cleanupPath(storageKey)
	if len(prefixs) == 0 {
		return nil, storage.ErrInvalidStorageKey
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storage.ErrFileNotFound
		}
		return nil, storage.ErrStorageError.WithDetail(err)
	}
	return &sectionReadCloser{ctx: ctx, reader: file, closer: file}, nil
}

func (l *localStorage) GetDirectURL(ctx context.Context, _ storage.StorageKey) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", storage.ErrDirectURLUnsupported
}

func (l *localStorage) Delete(ctx context.Context, storageKey storage.StorageKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	prefixs := l.cleanupPath(storageKey)
	if len(prefixs) == 0 {
		return storage.ErrStorageFailed
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)
	return os.Remove(path)
}

func (l *localStorage) cleanupPath(storageKey storage.StorageKey) []string {
	parts := strings.Split(storageKey.FileID(), "/")
	if len(parts) == 0 {
		return []string{}
	}
	cleaned := make([]string, 0)
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return []string{} // reject any unsafe segments
		}
		if !validFilename(part) {
			return []string{} // reject invalid segments
		}
		cleaned = append(cleaned, part)
	}
	return cleaned
}

func validFilename(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !unicode.IsDigit(r) && !unicode.IsLetter(r) && r != '_' && r != '-' && r != '.' && r != '~' {
			return false
		}
	}
	return true
}

func (l *localStorage) InitMultipartStore(ctx context.Context, storageKey storage.StorageKey) (string, error) {
	prefixs := l.cleanupPath(storageKey)
	if len(prefixs) == 0 {
		return "", storage.ErrInvalidStorageKey
	}
	uploadId := fmt.Sprintf("upload-%d", time.Now().UnixNano())
	l.uploadsLock.Lock()
	defer l.uploadsLock.Unlock()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	l.uploads[uploadId] = &uploadSession{
		storageKey: storageKey,
		uploadId:   uploadId,
		parts:      make(map[int]string),
		createdAt:  time.Now(),
	}
	return uploadId, nil
}

func (l *localStorage) StoreMultipart(ctx context.Context, uploadId string, partNumber int, data io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.uploadsLock.RLock()
	sess, ok := l.uploads[uploadId]
	l.uploadsLock.RUnlock()
	if !ok {
		return storage.ErrUploadSessionNotFound
	}

	partPath := filepath.Join(l.localPath, fmt.Sprintf(".%s.part-%d", uploadId, partNumber))
	f, err := os.CreateTemp(l.localPath, fmt.Sprintf(".%s.part-%d-*.tmp", uploadId, partNumber))
	if err != nil {
		return err
	}
	tempPath := f.Name()
	committed := false
	defer func() {
		_ = f.Close()
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()
	_, err = copyWithContext(ctx, f, data)
	if err != nil {
		if err := operationContextError(err); err != nil {
			return err
		}
		return storage.ErrStorageFailed.WithDetail(err)
	}
	if err := f.Close(); err != nil {
		return err
	}

	sess.partsMutex.Lock()
	defer sess.partsMutex.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, partPath); err != nil {
		return err
	}
	committed = true
	sess.parts[partNumber] = partPath
	return nil
}

func (l *localStorage) CompleteMultipart(ctx context.Context, uploadId string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.uploadsLock.RLock()
	sess, ok := l.uploads[uploadId]
	l.uploadsLock.RUnlock()
	if !ok {
		return storage.ErrUploadSessionNotFound
	}

	// Collect parts in order
	sess.partsMutex.Lock()
	defer sess.partsMutex.Unlock()
	partNumbers := make([]int, 0, len(sess.parts))
	for num := range sess.parts {
		partNumbers = append(partNumbers, num)
	}
	sort.Ints(partNumbers)

	prefixs := l.cleanupPath(sess.storageKey)
	if len(prefixs) == 0 {
		return storage.ErrStorageFailed
	}
	targetPath := filepath.Join(append([]string{l.localPath}, prefixs...)...)

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.CreateTemp(dir, ".multipart-*.tmp")
	if err != nil {
		return err
	}
	tempPath := f.Name()
	committed := false
	defer func() {
		_ = f.Close()
		if !committed {
			_ = os.Remove(tempPath)
		}
	}()
	if err := f.Chmod(0644); err != nil {
		return err
	}

	for _, partNum := range partNumbers {
		path := sess.parts[partNum]
		pf, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = copyWithContext(ctx, f, pf)
		_ = pf.Close()
		if err != nil {
			if err := operationContextError(err); err != nil {
				return err
			}
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		return err
	}
	committed = true
	for _, path := range sess.parts {
		_ = os.Remove(path)
	}

	l.uploadsLock.Lock()
	delete(l.uploads, uploadId)
	l.uploadsLock.Unlock()

	return nil
}

func (l *localStorage) AbortMultipart(ctx context.Context, uploadId string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.uploadsLock.Lock()
	sess, ok := l.uploads[uploadId]
	if !ok {
		l.uploadsLock.Unlock()
		return nil // already gone
	}
	delete(l.uploads, uploadId)
	l.uploadsLock.Unlock()

	sess.partsMutex.Lock()
	for _, path := range sess.parts {
		_ = os.Remove(path)
	}
	sess.partsMutex.Unlock()
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

type contextWriter struct {
	ctx    context.Context
	writer io.Writer
}

func (w *contextWriter) Write(p []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	return w.writer.Write(p)
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(&contextWriter{ctx: ctx, writer: dst}, &contextReader{ctx: ctx, reader: src})
}

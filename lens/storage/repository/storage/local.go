package storage

import (
	"errors"
	"fmt"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/storage"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type uploadSession struct {
	fileId     storage.FileID
	uploadId   string
	parts      map[int]string // partNumber -> file path
	createdAt  time.Time
	partsMutex sync.Mutex
}

type localStorage struct {
	name        string
	localPath   string
	urlPrefix   string
	log         logger.ILogger `aperture:""`
	uploads     map[string]*uploadSession
	uploadsLock sync.RWMutex
}

func (l *localStorage) ProviderName() string {
	return "local." + l.name
}

func (l *localStorage) HealthCheck() error {
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
	// Check read permission
	file, err := os.Open(l.localPath)
	if err == nil {
		_ = file.Close()
	} else if os.IsPermission(err) {
		return storage.ErrStorageError.WithDetailStr("no read permission")
	} else {
		return storage.ErrStorageError.WithDetail(err)
	}

	// Check write permission using a unique temp filename
	tempFilename := filepath.Join(l.localPath, fmt.Sprintf(".healthcheck_%s.tmp", strconv.FormatInt(time.Now().UnixMilli()/1000, 10)))

	for i := 0; i < 10; i++ {
		// First check if the file exists (extremely unlikely with our random name)
		if _, err := os.Stat(tempFilename); err == nil {
			// If by some miracle the file exists, generate a new name
			tempFilename = filepath.Join(l.localPath, fmt.Sprintf(".healthcheck_%s.tmp", strconv.FormatInt(time.Now().UnixMilli()/1000, 10)))
			fmt.Println(tempFilename)
		} else {
			break
		}
	}

	err = os.WriteFile(tempFilename, []byte("test"), 0644)
	if err != nil {
		return storage.ErrStorageError.WithDetailStr("no write permission")
	}
	_ = file.Close()
	if err := os.Remove(tempFilename); err != nil {
		return storage.ErrStorageError.WithDetailStr("no delete permission")
	}

	return nil
}

func (l *localStorage) Setup() error {
	l.log.Infof("local storage init with path: %s error: %v", l.localPath, l.HealthCheck())
	return nil
}

func NewLocalStorage(name string, localPath string) storage.IStorageProvider {
	return &localStorage{name: name, localPath: localPath, uploads: make(map[string]*uploadSession)}
}

func (l *localStorage) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IStorageProvider", "local")
}

func (l *localStorage) Meta(fileId storage.FileID) (storage.FileMeta, error) {
	pathParts := l.cleanupPath(fileId)
	if len(pathParts) == 0 {
		return storage.FileMeta{}, storage.ErrInvalidFileID
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
		FileID:           fileId,
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

func (l *localStorage) Store(fileId storage.FileID, data io.Reader) (err error) {
	prefixs := l.cleanupPath(fileId)
	if len(prefixs) == 0 {
		return storage.ErrStorageFailed
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return errors.New("failed to create directory")
	}
	file, err := os.Create(path)
	if err != nil {
		return errors.New("failed to create file")
	}
	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}
	return file.Close()
}

func (l *localStorage) Load(fileId storage.FileID, offset, length int64) (reader io.Reader, err error) {
	prefixs := l.cleanupPath(fileId)
	if len(prefixs) == 0 {
		return nil, storage.ErrStorageFailed
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)

	if offset < 0 {
		return nil, storage.ErrInvalidOffset
	}

	// If length is non-positive, read till EOF
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

	if offset > 0 {
		_, err := file.Seek(offset, io.SeekStart)
		if err != nil {
			_ = file.Close()
			return nil, storage.ErrStorageError.WithDetailStr("failed to seek in file")
		}
	}

	return io.LimitReader(file, length), nil
}

func (l *localStorage) LoadAll(fileId storage.FileID) (reader io.Reader, err error) {
	prefixs := l.cleanupPath(fileId)
	if len(prefixs) == 0 {
		return nil, storage.ErrInvalidFileID
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)
	return os.Open(path)
}

func (l *localStorage) GetPublicURL(fileId storage.FileID) (uri string, err error) {
	prefixs := strings.Split(fileId.ID(), "/")
	if len(prefixs) == 0 {
		return "", storage.ErrInvalidFileID
	}
	return url.JoinPath(l.urlPrefix, prefixs...)
}

func (l *localStorage) Delete(fileId storage.FileID) error {
	prefixs := l.cleanupPath(fileId)
	if len(prefixs) == 0 {
		return storage.ErrStorageFailed
	}
	path := filepath.Join(append([]string{l.localPath}, prefixs...)...)
	return os.Remove(path)
}

func (l *localStorage) cleanupPath(fileId storage.FileID) []string {
	parts := strings.Split(fileId.ID(), "/")
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

func (l *localStorage) InitMultipartStore(fileId storage.FileID) (string, error) {
	prefixs := l.cleanupPath(fileId)
	if len(prefixs) == 0 {
		return "", storage.ErrInvalidFileID
	}
	uploadId := fmt.Sprintf("upload-%d", time.Now().UnixNano())
	l.uploadsLock.Lock()
	defer l.uploadsLock.Unlock()
	l.uploads[uploadId] = &uploadSession{
		fileId:    fileId,
		uploadId:  uploadId,
		parts:     make(map[int]string),
		createdAt: time.Now(),
	}
	return uploadId, nil
}

func (l *localStorage) StorePart(uploadId string, partNumber int, data io.Reader) error {
	l.uploadsLock.RLock()
	sess, ok := l.uploads[uploadId]
	l.uploadsLock.RUnlock()
	if !ok {
		return errors.New("upload session not found")
	}

	tempFile := filepath.Join(l.localPath, fmt.Sprintf(".%s.part-%d", uploadId, partNumber))
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, data)
	if err != nil {
		return err
	}

	sess.partsMutex.Lock()
	sess.parts[partNumber] = tempFile
	sess.partsMutex.Unlock()
	return nil
}

func (l *localStorage) CompleteMultipartStore(uploadId string) error {
	l.uploadsLock.RLock()
	sess, ok := l.uploads[uploadId]
	l.uploadsLock.RUnlock()
	if !ok {
		return errors.New("upload session not found")
	}

	// Collect parts in order
	sess.partsMutex.Lock()
	partNumbers := make([]int, 0, len(sess.parts))
	for num := range sess.parts {
		partNumbers = append(partNumbers, num)
	}
	sort.Ints(partNumbers)

	prefixs := l.cleanupPath(sess.fileId)
	if len(prefixs) == 0 {
		return storage.ErrStorageFailed
	}
	targetPath := filepath.Join(append([]string{l.localPath}, prefixs...)...)

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		sess.partsMutex.Unlock()
		return err
	}

	f, err := os.Create(targetPath)
	if err != nil {
		sess.partsMutex.Unlock()
		return err
	}
	defer f.Close()

	for _, partNum := range partNumbers {
		path := sess.parts[partNum]
		pf, err := os.Open(path)
		if err != nil {
			f.Close()
			sess.partsMutex.Unlock()
			return err
		}
		_, err = io.Copy(f, pf)
		pf.Close()
		if err != nil {
			f.Close()
			sess.partsMutex.Unlock()
			return err
		}
		_ = os.Remove(path)
	}
	sess.partsMutex.Unlock()

	l.uploadsLock.Lock()
	delete(l.uploads, uploadId)
	l.uploadsLock.Unlock()

	return nil
}

func (l *localStorage) AbortMultipartStore(uploadId string) error {
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

package sessiontracker

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/storage"
	"sync"
)

type memoryUploadSessionTracker struct {
	sync.RWMutex
	sessions map[string]storage.UploadSession
}

func NewMemoryUploadSessionTracker() storage.IUploadSessionTracker {
	return &memoryUploadSessionTracker{sessions: make(map[string]storage.UploadSession)}
}

func (m *memoryUploadSessionTracker) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IUploadSessionTracker", "memory")
}

func (m *memoryUploadSessionTracker) Save(uploadId string, session storage.UploadSession) error {
	m.Lock()
	defer m.Unlock()
	m.sessions[uploadId] = session
	return nil
}

func (m *memoryUploadSessionTracker) Get(uploadId string) (storage.UploadSession, error) {
	m.RLock()
	defer m.RUnlock()
	sess, ok := m.sessions[uploadId]
	if !ok {
		return storage.UploadSession{}, storage.ErrUploadSessionNotFound
	}
	return sess, nil
}

func (m *memoryUploadSessionTracker) Delete(uploadId string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.sessions, uploadId)
	return nil
}

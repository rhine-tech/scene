package sessiontracker

import (
	"context"
	"sync"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/storage"
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

func (m *memoryUploadSessionTracker) Save(ctx context.Context, uploadId string, session storage.UploadSession) error {
	m.Lock()
	defer m.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	m.sessions[uploadId] = session
	return nil
}

func (m *memoryUploadSessionTracker) Get(ctx context.Context, uploadId string) (storage.UploadSession, error) {
	m.RLock()
	defer m.RUnlock()
	if err := ctx.Err(); err != nil {
		return storage.UploadSession{}, err
	}
	sess, ok := m.sessions[uploadId]
	if !ok {
		return storage.UploadSession{}, storage.ErrUploadSessionNotFound
	}
	return sess, nil
}

func (m *memoryUploadSessionTracker) Delete(ctx context.Context, uploadId string) error {
	m.Lock()
	defer m.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	delete(m.sessions, uploadId)
	return nil
}

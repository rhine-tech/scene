package factory

import (
	"github.com/rhine-tech/scene"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/repository/sessiontracker"
)

type SessionTrackerProvider scene.IModuleDependencyProvider[storageApi.IUploadSessionTracker]

type SessionTrackerRedis struct {
	Root string
}

func (l SessionTrackerRedis) Provide() storageApi.IUploadSessionTracker {
	return sessiontracker.NewRedisUploadSessionTracker()
}

type SessionTrackerMemory struct {
}

func (l SessionTrackerMemory) Provide() storageApi.IUploadSessionTracker {
	return sessiontracker.NewMemoryUploadSessionTracker()
}

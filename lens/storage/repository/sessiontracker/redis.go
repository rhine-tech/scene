package sessiontracker

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/storage"
	"time"
)

type redisUploadSessionTracker struct {
	redisDs  datasource.RedisDataSource `aperture:""`
	redisKey string
}

func (r *redisUploadSessionTracker) ImplName() scene.ImplName {
	return storage.Lens.ImplName("IUploadSessionTracker", "redis")
}

func NewRedisUploadSessionTracker() storage.IUploadSessionTracker {
	return &redisUploadSessionTracker{redisKey: "upload_session_"}
}

func (r *redisUploadSessionTracker) Save(uploadId string, session storage.UploadSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.redisDs.Set(context.Background(), r.redisKey+uploadId, string(data), 24*time.Hour)
}

func (r *redisUploadSessionTracker) Get(uploadId string) (storage.UploadSession, error) {
	res, err := r.redisDs.Get(context.Background(), r.redisKey+uploadId)
	if errors.Is(err, redis.Nil) {
		return storage.UploadSession{}, storage.ErrUploadSessionNotFound
	}
	if err != nil {
		return storage.UploadSession{}, err
	}
	var sess storage.UploadSession
	return sess, json.Unmarshal([]byte(res), &sess)
}

func (r *redisUploadSessionTracker) Delete(uploadId string) error {
	return r.redisDs.Delete(context.Background(), r.redisKey+uploadId)
}

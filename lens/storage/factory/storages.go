package factory

import (
	"time"

	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/infrastructure/config"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/repository/storage"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils/must"
)

type StorageProvider scene.IModuleDependencyProvider[storageApi.IStorageProvider]

type Local struct {
	Root string
}

func (l Local) Default() Local {
	cfg := registry.Use[config.IConfig](nil)
	return Local{
		Root: cfg.GetString("storage.local.root"),
	}
}

func (l Local) Provide() storageApi.IStorageProvider {
	return storage.NewLocalStorage("default", l.Root)
}

type S3 struct {
	Name            string
	Endpoint        string
	Region          string
	AccessKey       string
	SecretKey       string
	Bucket          string
	TempDir         string
	UseSSL          bool
	ForcePathStyle  bool
	PresignedURLTTL time.Duration
}

func (s S3) Default() S3 {
	cfg := registry.Use[config.IConfig](nil)
	presignedURLTTLSeconds := cfg.GetInt("storage.s3.presigned_url_ttl_seconds")
	presignedURLTTL := time.Duration(presignedURLTTLSeconds) * time.Second
	if presignedURLTTL <= 0 {
		presignedURLTTL = 15 * time.Minute
	}
	return S3{
		Name:            cfg.GetString("storage.s3.name"),
		Endpoint:        cfg.GetString("storage.s3.endpoint"),
		Region:          cfg.GetString("storage.s3.region"),
		AccessKey:       cfg.GetString("storage.s3.access_key"),
		SecretKey:       cfg.GetString("storage.s3.secret_key"),
		Bucket:          cfg.GetString("storage.s3.bucket"),
		TempDir:         cfg.GetString("storage.s3.temp_dir"),
		UseSSL:          cfg.GetBool("storage.s3.use_ssl"),
		ForcePathStyle:  cfg.GetBool("storage.s3.force_path_style"),
		PresignedURLTTL: presignedURLTTL,
	}
}

func (s S3) Provide() storageApi.IStorageProvider {
	if s.Name == "" {
		s.Name = "default"
	}
	return must.PMust(storage.NewS3StorageWithPresignedURLTTL(
		s.Endpoint,
		s.AccessKey,
		s.SecretKey,
		s.Bucket,
		s.Name,
		s.UseSSL,
		s.ForcePathStyle,
		s.Region,
		s.TempDir,
		s.PresignedURLTTL,
	))
}

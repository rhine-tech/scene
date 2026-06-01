package factory

import (
	"time"

	"github.com/rhine-tech/scene"
	storageApi "github.com/rhine-tech/scene/lens/storage"
	"github.com/rhine-tech/scene/lens/storage/repository/storage"
	"github.com/rhine-tech/scene/registry"
	"github.com/rhine-tech/scene/utils/must"
)

type StorageProvider scene.IModuleDependencyProvider[storageApi.IStorageProvider]

type Local struct {
	Root      string
	UrlPrefix string
}

func (l Local) Default() Local {
	return Local{
		Root:      registry.Config.GetString("storage.local.root"),
		UrlPrefix: registry.Config.GetString("storage.local.prefix"),
	}
}

func (l Local) Provide() storageApi.IStorageProvider {
	return registry.Load(storage.NewLocalStorage("default", l.Root, l.UrlPrefix))
}

type S3 struct {
	Name            string
	Endpoint        string
	Region          string
	AccessKey       string
	SecretKey       string
	Bucket          string
	UrlPrefix       string
	UseSSL          bool
	ForcePathStyle  bool
	DirectPublicURL bool
	PresignedURLTTL time.Duration
}

func (s S3) Default() S3 {
	presignedURLTTLSeconds := registry.Config.GetInt("storage.s3.presigned_url_ttl_seconds")
	presignedURLTTL := time.Duration(presignedURLTTLSeconds) * time.Second
	if presignedURLTTL <= 0 {
		presignedURLTTL = 15 * time.Minute
	}
	return S3{
		Name:            registry.Config.GetString("storage.s3.name"),
		Endpoint:        registry.Config.GetString("storage.s3.endpoint"),
		Region:          registry.Config.GetString("storage.s3.region"),
		AccessKey:       registry.Config.GetString("storage.s3.access_key"),
		SecretKey:       registry.Config.GetString("storage.s3.secret_key"),
		Bucket:          registry.Config.GetString("storage.s3.bucket"),
		UrlPrefix:       registry.Config.GetString("storage.s3.prefix"),
		UseSSL:          registry.Config.GetBool("storage.s3.use_ssl"),
		ForcePathStyle:  registry.Config.GetBool("storage.s3.force_path_style"),
		DirectPublicURL: registry.Config.GetBool("storage.s3.direct_public_url"),
		PresignedURLTTL: presignedURLTTL,
	}
}

func (s S3) Provide() storageApi.IStorageProvider {
	if s.Name == "" {
		s.Name = "default"
	}
	return registry.Load(must.PMust(storage.NewS3StorageWithPublicURLMode(
		s.Endpoint,
		s.AccessKey,
		s.SecretKey,
		s.Bucket,
		s.Name,
		s.UrlPrefix,
		s.UseSSL,
		s.ForcePathStyle,
		s.Region,
		s.DirectPublicURL,
		s.PresignedURLTTL,
	)))
}

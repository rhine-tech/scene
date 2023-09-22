package repository

import (
	"context"
	"github.com/rhine-tech/scene/drivers/caches"
	"github.com/rhine-tech/scene/drivers/repos"
	"github.com/rhine-tech/scene/lens/infrastructure/cache"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/infrastructure/logger"
	"github.com/rhine-tech/scene/lens/middlewares/permission"
	"go.mongodb.org/mongo-driver/bson"
)

type permDBModel struct {
	Owner       permission.PermOwner
	Permissions []string
}

type mongoImplCached struct {
	mgDrv *repos.MongoDatasourceCollection[permDBModel]
	cache *caches.Cache[[]string]
	log   logger.ILogger `aperture:""`
}

func NewPermissionMongoRepoCached(ds datasource.MongoDataSource, cache cache.ICache) permission.PermissionRepository {
	return &mongoImplCached{
		mgDrv: repos.UseMongoDatasourceCollection[permDBModel](ds, "permissions"),
		cache: caches.UseCache[[]string]("permissions", cache),
	}
}

func (m *mongoImplCached) Setup() error {
	m.log = m.log.WithPrefix(m.RepoImplName())
	return nil
}

func (m *mongoImplCached) RepoImplName() string {
	return "permission.repository.mongo.cached"
}

func (m *mongoImplCached) Status() error {
	return m.mgDrv.Status()
}

func (m *mongoImplCached) GetPermissions(owner string) []*permission.Permission {
	var permStrs []string
	if val, exist := m.cache.Get(owner); exist {
		permStrs = val
	} else {
		result, err := m.mgDrv.FindOne(bson.M{"owner": owner})
		if err != nil {
			return []*permission.Permission{}
		}
		permStrs = result.Permissions
		_ = m.cache.Set(owner, permStrs)
	}
	var permissions []*permission.Permission
	for _, perm := range permStrs {
		p, _ := permission.ParsePermission(perm)
		permissions = append(permissions, p)
	}
	return permissions
}

func (m *mongoImplCached) AddPermission(owner string, perm string) (*permission.Permission, error) {
	p, err := permission.ParsePermission(perm)
	if err != nil {
		return nil, err
	}

	_, err = m.mgDrv.Collection().UpdateOne(
		context.Background(),
		bson.M{"owner": owner},
		bson.M{"$addToSet": bson.M{"permissions": perm}},
	)
	if err != nil {
		return nil, err
	}

	if err := m.cache.Delete(owner); err != nil {
		m.log.Errorf("remove permission from cache failed: %s", err.Error())
	}

	return p, nil
}

func (m *mongoImplCached) RemovePermission(owner string, perm string) error {
	_, err := m.mgDrv.Collection().UpdateOne(
		context.Background(),
		bson.M{"owner": owner},
		bson.M{"$pull": bson.M{"permissions": perm}},
	)

	if err := m.cache.Delete(owner); err != nil {
		m.log.Errorf("remove permission from cache failed: %s", err.Error())
	}

	return err
}

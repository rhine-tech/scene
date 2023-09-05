package repository

import (
	"context"
	"github.com/rhine-tech/scene/drivers/repos"
	"github.com/rhine-tech/scene/lens/middlewares/permission"
	"github.com/rhine-tech/scene/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoImpl struct {
	*repos.MongoRepo
	collection *mongo.Collection
}

func (m *mongoImpl) RepoImplName() string {
	return "permission.repository.mongo"
}

func NewPermissionMongoRepo(cfg model.DatabaseConfig) permission.PermissionRepository {
	repo := &mongoImpl{
		MongoRepo: repos.NewMongoRepo(cfg),
	}
	if repo.Status() != nil {
		return repo
	}
	repo.collection = repo.Database().Collection("permissions")
	return repo
}

func (m *mongoImpl) GetPermissions(owner string) []*permission.Permission {
	var result struct {
		Owner       permission.PermOwner
		Permissions []string
	}
	err := m.collection.FindOne(context.Background(), bson.M{"owner": owner}).Decode(&result)
	if err != nil {
		return []*permission.Permission{}
	}

	var permissions []*permission.Permission
	for _, perm := range result.Permissions {
		p, _ := permission.ParsePermission(perm)
		permissions = append(permissions, p)
	}

	return permissions
}

func (m *mongoImpl) AddPermission(owner string, perm string) (*permission.Permission, error) {
	p, err := permission.ParsePermission(perm)
	if err != nil {
		return nil, err
	}

	_, err = m.collection.UpdateOne(
		context.Background(),
		bson.M{"owner": owner},
		bson.M{"$addToSet": bson.M{"permissions": perm}},
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (m *mongoImpl) RemovePermission(owner string, perm string) error {
	_, err := m.collection.UpdateOne(
		context.Background(),
		bson.M{"owner": owner},
		bson.M{"$pull": bson.M{"permissions": perm}},
	)
	return err
}

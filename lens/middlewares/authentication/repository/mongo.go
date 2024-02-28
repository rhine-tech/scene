package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/lens/infrastructure/datasource"
	"github.com/rhine-tech/scene/lens/middlewares/authentication"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoImpl struct {
	ds         datasource.MongoDataSource `aperture:""`
	collection *mongo.Collection
}

func (m *mongoImpl) RepoImplName() scene.ImplName {
	return scene.NewRepoImplName("authentication", "AuthenticationManageRepository", "mongo")
}

func (m *mongoImpl) Status() error {
	return m.ds.Status()
}

func (m *mongoImpl) Setup() error {
	m.collection = m.ds.Collection("users")
	return nil
}

func NewMongoAuthenticationRepository(ds datasource.MongoDataSource) authentication.AuthenticationManageRepository {
	repo := &mongoImpl{
		ds: ds,
	}
	return repo
}

func (m *mongoImpl) queryBy(filter bson.M) (authentication.User, error) {
	var result authentication.User
	err := m.collection.FindOne(context.Background(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return result, authentication.ErrUserNotFound
	}
	return result, err
}

func (m *mongoImpl) UserById(userId string) (authentication.User, error) {
	return m.queryBy(bson.M{"user_id": userId})
}

func (m *mongoImpl) UserByName(username string) (authentication.User, error) {
	return m.queryBy(bson.M{"username": username})
}

func (m *mongoImpl) UserByEmail(email string) (authentication.User, error) {
	return m.queryBy(bson.M{"email": email})
}

func (m *mongoImpl) Authenticate(username string, password string) (string, error) {
	var user authentication.User
	err := m.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)

	if err != nil || user.Password != password {
		return "", err
	}

	return user.UserID, nil
}

func (m *mongoImpl) AddUser(username, password string) (authentication.User, error) {
	var user authentication.User
	err := m.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)

	if err == nil {
		return authentication.User{}, authentication.ErrUserAlreadyExists
	}

	userID := uuid.New().String()
	newUser := &authentication.User{
		UserID:   userID,
		Username: username,
		Password: password,
	}

	_, err = m.collection.InsertOne(context.Background(), newUser)

	if err != nil {
		return authentication.User{}, err
	}

	return *newUser, nil
}

func (m *mongoImpl) DeleteUser(userID string) error {
	deleteResult, err := m.collection.DeleteOne(context.Background(), bson.M{"user_id": userID})

	if err != nil || deleteResult.DeletedCount == 0 {
		return authentication.ErrUserNotFound
	}

	return nil
}

func (m *mongoImpl) UpdateUser(user authentication.User) error {
	updateResult, err := m.collection.UpdateOne(
		context.Background(),
		bson.M{"username": user.Username},
		bson.M{"$set": bson.M{"username": user.Username, "password": user.Password}},
	)

	if err != nil || updateResult.ModifiedCount == 0 {
		return authentication.ErrUserNotFound
	}

	return nil
}

type mongoInfoImpl struct {
	ds         datasource.MongoDataSource `aperture:""`
	collection *mongo.Collection
}

func (m *mongoInfoImpl) RepoImplName() scene.ImplName {
	return scene.NewRepoImplName("authentication", "UserInfo", "mongo")
}

func (m *mongoInfoImpl) Status() error {
	return m.ds.Status()
}

func (m *mongoInfoImpl) Setup() error {
	m.collection = m.ds.Collection("user_info")
	return m.ds.Status()
}

func NewUserInfoRepository(ds datasource.MongoDataSource) authentication.UserInfoRepository {
	repo := &mongoInfoImpl{
		ds: ds,
	}
	return repo
}

func (m mongoInfoImpl) InfoById(userId string) (authentication.UserInfo, error) {
	var result authentication.UserInfo
	err := m.collection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&result)
	return result, err
}

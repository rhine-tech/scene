package repository

import (
	"context"
	"github.com/aynakeya/scene/drivers/repos"
	"github.com/aynakeya/scene/lens/middlewares/authentication"
	"github.com/aynakeya/scene/model"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoImpl struct {
	*repos.MongoRepo
	collection *mongo.Collection
}

func NewMongoAuthenticationRepository(cfg model.DatabaseConfig) authentication.AuthenticationManageRepository {
	repo := &mongoImpl{
		MongoRepo: repos.NewMongoRepo(cfg),
	}
	if repo.Status() != nil {
		return repo
	}
	repo.collection = repo.Database().Collection("users")
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
		return "", authentication.ErrAuthenticationFailed
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
	*repos.MongoRepo
	collection *mongo.Collection
}

func (m *mongoImpl) RepoImplName() string {
	return "authentication.repository.mongo"
}

func NewUserInfoRepository(cfg model.DatabaseConfig) authentication.UserInfoRepository {
	repo := &mongoInfoImpl{
		MongoRepo: repos.NewMongoRepo(cfg),
	}
	if repo.Status() != nil {
		return repo
	}
	repo.collection = repo.Database().Collection("user_info")
	return repo
}

func (m mongoInfoImpl) InfoById(userId string) (authentication.UserInfo, error) {
	var result authentication.UserInfo
	err := m.collection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&result)
	return result, err
}

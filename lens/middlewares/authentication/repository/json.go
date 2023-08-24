package repository

import (
	"encoding/json"
	"github.com/aynakeya/scene/lens/middlewares/authentication"
	"github.com/google/uuid"
	"os"
)

type JSONAuthenticationRepository struct {
	cfg   authentication.AuthRepoConfig
	users map[string]*authentication.User
	err   error
}

func (jar *JSONAuthenticationRepository) RepoImplName() string {
	return "json"
}

func (jar *JSONAuthenticationRepository) Status() error {
	return jar.err
}

//func NewJSONAuthenticationRepository(cfg authentication.AuthRepoConfig) authentication.AuthenticationManageRepository {
//	tmp := &JSONAuthenticationRepository{
//		cfg:   cfg,
//		users: make(map[string]*authentication.User),
//	}
//	tmp.err = tmp.load()
//	return tmp
//}

func (j *JSONAuthenticationRepository) load() error {
	data, err := os.ReadFile(j.cfg.SaveDir)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &j.users)
}

func (jar *JSONAuthenticationRepository) Authenticate(username string, password string) (string, error) {
	user, ok := jar.users[username]
	if !ok || user.Password != password {
		return "", authentication.ErrAuthenticationFailed
	}
	return user.UserID, nil
}

func (jar *JSONAuthenticationRepository) GetUser(username string) (authentication.User, bool) {
	user, ok := jar.users[username]
	if !ok {
		return authentication.User{}, false
	}
	return *user, true
}

func (jar *JSONAuthenticationRepository) AddUser(username, password string) (authentication.User, error) {
	if _, exists := jar.users[username]; exists {
		return authentication.User{}, authentication.ErrUserAlreadyExists
	}

	id := uuid.New().String()
	for _, exists := jar.users[id]; exists; {
		id = uuid.New().String()
	}

	newUser := &authentication.User{
		UserID:   id,
		Username: username,
		Password: password,
	}
	jar.users[username] = newUser
	return *newUser, nil
}

func (jar *JSONAuthenticationRepository) DeleteUser(userID string) error {
	for username, user := range jar.users {
		if user.UserID == userID {
			delete(jar.users, username)
			return nil
		}
	}
	return authentication.ErrUserNotFound
}

func (jar *JSONAuthenticationRepository) UpdateUser(user authentication.User) error {
	_, ok := jar.users[user.Username]
	if !ok {
		return authentication.ErrUserNotFound
	}
	jar.users[user.Username] = &user
	return nil
}

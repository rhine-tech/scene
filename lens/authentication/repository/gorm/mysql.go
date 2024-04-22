package gorm

import (
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/spf13/cast"
	"strconv"
)

type mysqlImpl struct {
	gorm orm.Gorm `aperture:""`
}

func AuthenticationRepository(gorm orm.Gorm) authentication.AuthenticationRepository {
	return &mysqlImpl{gorm: gorm}
}

func (m *mysqlImpl) Setup() error {
	err := m.gorm.RegisterModel(&tableUser{}, &tableUserInfo{})
	if err != nil {
		return err
	}
	return nil
}

func (m *mysqlImpl) RepoImplName() scene.ImplName {
	return authentication.Lens.ImplName("AuthenticationRepository", "mysql")
}

// Authenticate checks if the username and password match a user in the database
func (m *mysqlImpl) Authenticate(username string, password string) (string, error) {
	var user tableUser
	err := m.gorm.DB().Where("username = ? AND password = ?", username, password).First(&user).Error
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(user.UserID, 10), nil // Convert UserID from uint64 to string
}

// UserById finds a user by their user ID
func (m *mysqlImpl) UserById(userId string) (authentication.User, error) {
	var user tableUser
	id, err := strconv.ParseUint(userId, 10, 64) // Convert userId from string to uint64
	if err != nil {
		return authentication.User{}, err
	}
	err = m.gorm.DB().First(&user, "user_id = ?", id).Error
	return user.toUser(), err
}

// UserByName finds a user by their username
func (m *mysqlImpl) UserByName(username string) (authentication.User, error) {
	var user tableUser
	err := m.gorm.DB().First(&user, "username = ?", username).Error
	return user.toUser(), err
}

// UserByEmail finds a user by their email
func (m *mysqlImpl) UserByEmail(email string) (authentication.User, error) {
	var user tableUser
	err := m.gorm.DB().First(&user, "email = ?", email).Error
	return user.toUser(), err
}

// AddUser creates a new user in the database
func (m *mysqlImpl) AddUser(username, password string) (authentication.User, error) {
	user := tableUser{
		Username: username,
		Password: password,
		Info: tableUserInfo{
			DisplayName: username,
		},
	}
	err := m.gorm.DB().Create(&user).Error
	if err != nil {
		return authentication.User{}, err
	}
	return authentication.User{
		UserID:   strconv.FormatUint(user.UserID, 10),
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// DeleteUser removes a user by their user ID
func (m *mysqlImpl) DeleteUser(userId string) error {
	id, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		return err
	}
	return m.gorm.DB().Delete(&tableUser{}, id).Error
}

// UpdateUser updates a user's details in the database
func (m *mysqlImpl) UpdateUser(user authentication.User) error {
	updatedUser := tableUser{
		UserID:   cast.ToUint64(user.UserID),
		Username: user.Username,
		Password: user.Password, // Assumes password may also need to be updated
		Email:    user.Email,
	}
	return m.gorm.DB().Save(&updatedUser).Error
}

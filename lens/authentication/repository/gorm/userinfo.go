package gorm

import (
	"errors"
	"fmt"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/authentication"
	"gorm.io/gorm"
	"strconv"
)

// mysqlUserInfoImpl is the GORM implementation of UserInfoRepository
type mysqlUserInfoImpl struct {
	gorm orm.Gorm `aperture:""`
}

func (m *mysqlUserInfoImpl) RepoImplName() scene.ImplName {
	return authentication.Lens.ImplName("UserInfoRepository", "mysql")
}

// NewUserInfoRepository initializes a new instance of UserInfoRepository with GORM
func NewUserInfoRepository(gorm orm.Gorm) authentication.UserInfoRepository {
	return &mysqlUserInfoImpl{gorm: gorm}
}

// InfoById fetches user info by user ID
func (m *mysqlUserInfoImpl) InfoById(userId string) (authentication.UserInfo, error) {
	var info tableUser
	uid, err := strconv.ParseUint(userId, 10, 64) // Convert string ID to uint64
	if err != nil {
		return authentication.UserInfo{}, fmt.Errorf("invalid user ID: %v", err)
	}

	err = m.gorm.DB().Preload("Info").First(&info, "user_id = ?", uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return authentication.UserInfo{}, authentication.ErrUserNotFound.WithDetailStr(userId)
		}
		return authentication.UserInfo{}, err
	}

	// Assuming you have a function to map tableUserInfo to UserInfo
	return authentication.UserInfo{
		UserID:      strconv.FormatUint(info.UserID, 10),
		DisplayName: info.Info.DisplayName,
		Avatar:      info.Info.Avatar,
		Username:    info.Username,
		Email:       info.Email,
	}, nil
}

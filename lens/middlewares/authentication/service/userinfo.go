package service

import "github.com/aynakeya/scene/lens/middlewares/authentication"

type userInfoServiceImpl struct {
	authRepo authentication.AuthenticationRepository
	infoRepo authentication.UserInfoRepository
}

func (u *userInfoServiceImpl) SrvImplName() string {
	return "authentication.service.userinfo.v1"
}

func (u *userInfoServiceImpl) InfoById(userId string) (authentication.UserInfo, error) {
	user, err := u.authRepo.UserById(userId)
	if err != nil {
		return authentication.UserInfo{}, err
	}
	info, _ := u.infoRepo.InfoById(userId)
	info.UserID = user.UserID
	info.Username = user.Username
	info.Email = user.Email
	if info.DisplayName == "" {
		info.DisplayName = user.Username
	}
	return info, nil
}

func NewUserInfoService(authRepo authentication.AuthenticationRepository, infoRepo authentication.UserInfoRepository) authentication.UserInfoService {
	return &userInfoServiceImpl{
		authRepo: authRepo,
		infoRepo: infoRepo,
	}
}

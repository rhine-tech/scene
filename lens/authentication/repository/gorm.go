package repository

import (
	"errors"
	"github.com/rhine-tech/scene"
	"github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
	"github.com/rhine-tech/scene/model/query"
	"gorm.io/gorm"
)

// gormAuthRepo 实现了 IAuthenticationRepository
type gormAuthRepo struct {
	*orm.GormRepository[authentication.User] `aperture:""`
}

// NewGormAuthenticationRepository 创建一个 IAuthenticationRepository 的 GORM 实现
func NewGormAuthenticationRepository(db orm.Gorm) authentication.IAuthenticationRepository {
	return &gormAuthRepo{
		GormRepository: orm.NewGormRepository[authentication.User](db, make(query.FieldMapper)),
	}
}

func (r *gormAuthRepo) ImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationRepository", "gorm")
}

func (r *gormAuthRepo) Authenticate(username, password string) (string, error) {
	user, found, err := r.FindFirst(
		query.Field("username").Equal(username),
		query.Field("password").Equal(password))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if !found {
		return "", authentication.ErrAuthenticationFailed
	}
	return user.UserID, nil
}

func (r *gormAuthRepo) UserById(userId string) (authentication.User, error) {
	user, found, err := r.FindFirst(query.Field("user_id").Equal(userId))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return authentication.User{}, err
	}
	if !found {
		return authentication.User{}, authentication.ErrUserNotFound
	}
	return user, nil
}

func (r *gormAuthRepo) UserByName(username string) (authentication.User, error) {
	user, found, err := r.FindFirst(query.Field("username").Equal(username))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return authentication.User{}, err
	}
	if !found {
		return authentication.User{}, authentication.ErrUserNotFound
	}
	return user, nil
}

func (r *gormAuthRepo) UserByEmail(email string) (authentication.User, error) {
	user, found, err := r.FindFirst(query.Field("email").Equal(email))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return authentication.User{}, err
	}
	if !found {
		return authentication.User{}, authentication.ErrUserNotFound
	}
	return user, nil
}

func (r *gormAuthRepo) AddUser(user authentication.User) (authentication.User, error) {
	// 为保证唯一性，先检查用户是否存在
	_, found, err := r.FindFirst(query.Field("username").Equal(user.Username))
	if err != nil {
		return authentication.User{}, err
	}
	if found {
		return authentication.User{}, authentication.ErrUserAlreadyExists
	}
	if err := r.Create(&user); err != nil {
		return authentication.User{}, err
	}
	return user, nil
}

func (r *gormAuthRepo) DeleteUser(userId string) error {
	// GormRepository 的 Delete 需要一个 Option 来指定删除条件
	return r.Delete(query.Field("user_id").Equal(userId))
}

func (r *gormAuthRepo) UpdateUser(user authentication.User) error {
	// 将需要更新的字段放入 map 中
	updates := map[string]interface{}{
		"username":     user.Username,
		"password":     user.Password,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"timezone":     user.Timezone,
	}
	// 使用 Option 指定更新条件
	return r.Update(updates, query.Field("user_id").Equal(user.UserID))
}

// --- IAccessTokenRepository 实现 ---

// gormAccessTokenRepo 实现了 IAccessTokenRepository
type gormAccessTokenRepo struct {
	*orm.GormRepository[authentication.AccessToken] `aperture:""`
}

// NewGormAccessTokenRepository 创建一个 IAccessTokenRepository 的 GORM 实现
func NewGormAccessTokenRepository(db orm.Gorm) authentication.IAccessTokenRepository {
	return &gormAccessTokenRepo{
		GormRepository: orm.NewGormRepository[authentication.AccessToken](db, make(query.FieldMapper)),
	}
}

func (r *gormAccessTokenRepo) ImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAccessTokenRepository", "gorm")
}

func (r *gormAccessTokenRepo) CreateToken(token authentication.AccessToken) (authentication.AccessToken, error) {
	if err := r.Create(&token); err != nil {
		return authentication.AccessToken{}, err
	}
	return token, nil
}

func (r *gormAccessTokenRepo) GetTokenByValue(tokenValue string) (authentication.AccessToken, error) {
	token, found, err := r.FindFirst(query.Field("token").Equal(tokenValue))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return authentication.AccessToken{}, err
	}
	if !found {
		return authentication.AccessToken{}, authentication.ErrTokenNotFound
	}
	return token, nil
}

func (r *gormAccessTokenRepo) ListTokensByUser(userId string, offset, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	return r.List(offset, limit, query.Field("user_id").Equal(userId))
}

func (r *gormAccessTokenRepo) ListTokens(offset, limit int64) (model.PaginationResult[authentication.AccessToken], error) {
	return r.List(offset, limit)
}

func (r *gormAccessTokenRepo) DeleteToken(token string) error {
	return r.Delete(query.Field("token").Equal(token))
}

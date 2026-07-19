package repository

import (
	"context"
	"errors"

	"github.com/rhine-tech/scene"
	sceneorm "github.com/rhine-tech/scene/composition/orm"
	"github.com/rhine-tech/scene/lens/authentication"
	"github.com/rhine-tech/scene/model"
	"gorm.io/gorm"
)

type userRow struct {
	UserID      string `gorm:"column:user_id;primaryKey"`
	Username    string `gorm:"column:username;uniqueIndex"`
	Password    string `gorm:"column:password"`
	Email       string `gorm:"column:email"`
	DisplayName string `gorm:"column:display_name"`
	Avatar      string `gorm:"column:avatar"`
	Timezone    string `gorm:"column:timezone"`
}

func (userRow) TableName() string {
	return authentication.Lens.TableName("users")
}

func userRowFromDomain(user authentication.User) userRow {
	return userRow{
		UserID:      user.UserID,
		Username:    user.Username,
		Password:    user.Password,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Timezone:    user.Timezone,
	}
}

func (r userRow) toDomain() authentication.User {
	return authentication.User{
		UserID:      r.UserID,
		Username:    r.Username,
		Password:    r.Password,
		Email:       r.Email,
		DisplayName: r.DisplayName,
		Avatar:      r.Avatar,
		Timezone:    r.Timezone,
	}
}

type accessTokenRow struct {
	Token     string `gorm:"column:token;primaryKey"`
	UserID    string `gorm:"column:user_id"`
	Name      string `gorm:"column:name"`
	CreatedAt int64  `gorm:"column:created_at"`
	ExpireAt  int64  `gorm:"column:expire_at"`
}

func (accessTokenRow) TableName() string {
	return authentication.Lens.TableName("access_tokens")
}

func accessTokenRowFromDomain(token authentication.AccessToken) accessTokenRow {
	return accessTokenRow{
		Token:     token.Token,
		UserID:    token.UserID,
		Name:      token.Name,
		CreatedAt: token.CreatedAt,
		ExpireAt:  token.ExpireAt,
	}
}

func (r accessTokenRow) toDomain() authentication.AccessToken {
	return authentication.AccessToken{
		Token:     r.Token,
		UserID:    r.UserID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
		ExpireAt:  r.ExpireAt,
	}
}

type gormAuthRepo struct {
	db *sceneorm.Gorm `aperture:""`
}

func NewGormAuthenticationRepository(db *sceneorm.Gorm) authentication.IAuthenticationRepository {
	return &gormAuthRepo{db: db}
}

func (r *gormAuthRepo) Setup() error {
	return r.db.AutoMigrate(&userRow{})
}

func (r *gormAuthRepo) ImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAuthenticationRepository", "gorm")
}

func (r *gormAuthRepo) Authenticate(ctx context.Context, username, password string) (string, error) {
	var row userRow
	err := r.db.Session(ctx).
		Where(map[string]any{"username": username, "password": password}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", authentication.ErrAuthenticationFailed
	}
	if err != nil {
		return "", err
	}
	return row.UserID, nil
}

func (r *gormAuthRepo) UserById(ctx context.Context, userID string) (authentication.User, error) {
	return r.findUser(ctx, map[string]any{"user_id": userID})
}

func (r *gormAuthRepo) UserByName(ctx context.Context, username string) (authentication.User, error) {
	return r.findUser(ctx, map[string]any{"username": username})
}

func (r *gormAuthRepo) UserByEmail(ctx context.Context, email string) (authentication.User, error) {
	return r.findUser(ctx, map[string]any{"email": email})
}

func (r *gormAuthRepo) findUser(ctx context.Context, where map[string]any) (authentication.User, error) {
	var row userRow
	err := r.db.Session(ctx).Where(where).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authentication.User{}, authentication.ErrUserNotFound
	}
	if err != nil {
		return authentication.User{}, err
	}
	return row.toDomain(), nil
}

func (r *gormAuthRepo) AddUser(ctx context.Context, user authentication.User) (authentication.User, error) {
	row := userRowFromDomain(user)
	err := r.db.Session(ctx).Create(&row).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return authentication.User{}, authentication.ErrUserAlreadyExists
	}
	if err != nil {
		return authentication.User{}, err
	}
	return row.toDomain(), nil
}

func (r *gormAuthRepo) DeleteUser(ctx context.Context, userID string) error {
	return r.db.Session(ctx).
		Where(&userRow{UserID: userID}).
		Delete(&userRow{}).Error
}

func (r *gormAuthRepo) UpdateUser(ctx context.Context, user authentication.User) error {
	updates := map[string]any{
		"username":     user.Username,
		"password":     user.Password,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"avatar":       user.Avatar,
		"timezone":     user.Timezone,
	}
	return r.db.Session(ctx).
		Model(&userRow{}).
		Where(&userRow{UserID: user.UserID}).
		Updates(updates).Error
}

func (r *gormAuthRepo) ListUsers(
	ctx context.Context,
	offset, limit int64,
) (model.PaginationResult[authentication.User], error) {
	result := model.PaginationResult[authentication.User]{
		Offset:  offset,
		Results: make([]authentication.User, 0),
	}

	if err := r.db.Session(ctx).Model(&userRow{}).Count(&result.Total).Error; err != nil {
		return result, err
	}

	var rows []userRow
	if err := r.db.Session(ctx).
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&rows).Error; err != nil {
		return result, err
	}

	result.Results = make([]authentication.User, len(rows))
	for i, row := range rows {
		result.Results[i] = row.toDomain()
	}
	result.Count = int64(len(result.Results))
	return result, nil
}

type gormAccessTokenRepo struct {
	db *sceneorm.Gorm `aperture:""`
}

func NewGormAccessTokenRepository(db *sceneorm.Gorm) authentication.IAccessTokenRepository {
	return &gormAccessTokenRepo{db: db}
}

func (r *gormAccessTokenRepo) Setup() error {
	return r.db.AutoMigrate(&accessTokenRow{})
}

func (r *gormAccessTokenRepo) ImplName() scene.ImplName {
	return authentication.Lens.ImplName("IAccessTokenRepository", "gorm")
}

func (r *gormAccessTokenRepo) CreateToken(
	ctx context.Context,
	token authentication.AccessToken,
) (authentication.AccessToken, error) {
	row := accessTokenRowFromDomain(token)
	if err := r.db.Session(ctx).Create(&row).Error; err != nil {
		return authentication.AccessToken{}, err
	}
	return row.toDomain(), nil
}

func (r *gormAccessTokenRepo) GetTokenByValue(
	ctx context.Context,
	tokenValue string,
) (authentication.AccessToken, error) {
	var row accessTokenRow
	err := r.db.Session(ctx).
		Where(&accessTokenRow{Token: tokenValue}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authentication.AccessToken{}, authentication.ErrTokenNotFound
	}
	if err != nil {
		return authentication.AccessToken{}, err
	}
	return row.toDomain(), nil
}

func (r *gormAccessTokenRepo) ListTokensByUser(
	ctx context.Context,
	userID string,
	offset, limit int64,
) (model.PaginationResult[authentication.AccessToken], error) {
	return r.listTokens(ctx, offset, limit, &accessTokenRow{UserID: userID})
}

func (r *gormAccessTokenRepo) ListTokens(
	ctx context.Context,
	offset, limit int64,
) (model.PaginationResult[authentication.AccessToken], error) {
	return r.listTokens(ctx, offset, limit, nil)
}

func (r *gormAccessTokenRepo) listTokens(
	ctx context.Context,
	offset, limit int64,
	where *accessTokenRow,
) (model.PaginationResult[authentication.AccessToken], error) {
	result := model.PaginationResult[authentication.AccessToken]{
		Offset:  offset,
		Results: make([]authentication.AccessToken, 0),
	}

	countQuery := r.db.Session(ctx).Model(&accessTokenRow{})
	findQuery := r.db.Session(ctx)
	if where != nil {
		countQuery = countQuery.Where(where)
		findQuery = findQuery.Where(where)
	}
	if err := countQuery.Count(&result.Total).Error; err != nil {
		return result, err
	}

	var rows []accessTokenRow
	if err := findQuery.
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&rows).Error; err != nil {
		return result, err
	}

	result.Results = make([]authentication.AccessToken, len(rows))
	for i, row := range rows {
		result.Results[i] = row.toDomain()
	}
	result.Count = int64(len(result.Results))
	return result, nil
}

func (r *gormAccessTokenRepo) DeleteToken(ctx context.Context, tokenValue string) error {
	return r.db.Session(ctx).
		Where(&accessTokenRow{Token: tokenValue}).
		Delete(&accessTokenRow{}).Error
}

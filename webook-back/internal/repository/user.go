package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/repository/cache"
	"github.com/ac-zht/webook/internal/repository/dao"
	"time"
)

var ErrUserDuplicate = dao.ErrUserDuplicate
var ErrUserNotFound = errors.New("用户未找到")

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(d dao.UserDAO, c cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   d,
		cache: c,
	}
}

func (ur *CachedUserRepository) Create(ctx context.Context, u domain.User) error {
	du := ur.domainToEntity(u)
	return ur.dao.Insert(ctx, du)
}

func (ur *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := ur.cache.Get(ctx, id)
	if err == nil {
		return u, nil
	}
	ue, err := ur.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	u = ur.entityToDomain(ue)
	_ = ur.cache.Set(ctx, u)
	return u, nil
}

func (ur *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return ur.entityToDomain(u), nil
}

func (ur *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := ur.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return ur.entityToDomain(u), nil
}

func (ur *CachedUserRepository) Update(ctx context.Context, u domain.User) error {
	err := ur.dao.UpdateNonZeroFields(ctx, ur.domainToEntity(u))
	if err != nil {
		return err
	}
	return nil
}

func (ur *CachedUserRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Password: user.Password,
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Nickname: sql.NullString{
			String: user.Nickname,
			Valid:  user.Nickname != "",
		},
		Birthday: sql.NullInt64{
			Int64: user.Birthday.UnixMilli(),
			Valid: !user.Birthday.IsZero(),
		},
		AboutMe: sql.NullString{
			String: user.AboutMe,
			Valid:  user.AboutMe != "",
		},
	}
}

func (ur *CachedUserRepository) entityToDomain(user dao.User) domain.User {
	var birthday time.Time
	if user.Birthday.Valid {
		birthday = time.UnixMilli(user.Birthday.Int64)
	}
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password,
		Phone:    user.Phone.String,
		Nickname: user.Nickname.String,
		Birthday: birthday,
		AboutMe:  user.AboutMe.String,
		Ctime:    time.UnixMilli(user.Ctime),
	}
}

package repository

import (
	"context"
	"errors"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository/dao"
	"time"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicate
var ErrUserNotFound = errors.New("用户未找到")

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(d *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: d,
	}
}

func (ur *UserRepository) Create(ctx context.Context, u domain.User) error {
	return ur.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}, nil
}

func (ur *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := ur.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
		AboutMe:  u.AboutMe,
		Password: u.Password,
		Ctime:    time.UnixMilli(u.Ctime),
	}, nil
}

func (ur *UserRepository) Update(ctx context.Context, u domain.User) error {
	err := ur.dao.UpdateNonZeroFields(ctx, ur.domainToEntity(u))
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Birthday: user.Birthday.UnixMilli(),
		AboutMe:  user.AboutMe,
	}
}

func (ur *UserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Birthday: time.UnixMilli(user.Birthday),
		AboutMe:  user.AboutMe,
		Ctime:    time.UnixMilli(user.Ctime),
	}
}

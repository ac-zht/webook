package service

import (
	"context"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (srv *UserService) Signup(ctx context.Context, u domain.User) error {
	return srv.repo.Create(ctx, u)
}

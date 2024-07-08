package repository

import (
	"context"
	"github.com/ac-zht/webook/internal/repository/cache"
)

var (
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
)

type CodeRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func (c CachedCodeRepository) Store(ctx context.Context, biz string, phone string, code string) error {
	return c.cache.Set(ctx, biz, phone, code)
}

func (c CachedCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, inputCode)
}

func NewCacheCodeRepository(c cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: c,
	}
}

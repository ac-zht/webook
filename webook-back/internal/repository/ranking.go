package repository

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	redisCache *cache.RedisRankingCache
	localCache *cache.RankingLocalCache
	topN       atomicx.Value[[]domain.Article]
}

func NewCachedRankingRepository(
	redisCache *cache.RedisRankingCache,
	localCache *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepository{
		redisCache: redisCache,
		localCache: localCache,
	}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.localCache.Get(ctx)
	if err == nil {
		return arts, nil
	}
	arts, err = c.redisCache.Get(ctx)
	if err == nil {
		_ = c.localCache.Set(ctx, arts)
		return arts, err
	}
	return c.localCache.ForceGet(ctx)
}

package cache

import (
	"context"
	"encoding/json"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RedisRankingCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewRedisRankingCache(client redis.Cmdable) *RedisRankingCache {
	return &RedisRankingCache{
		key:        "ranking:article",
		client:     client,
		expiration: time.Minute * 3,
	}
}

func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func (r *RedisRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return nil, err
}

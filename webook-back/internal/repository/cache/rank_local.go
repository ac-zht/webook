package cache

import (
	"context"
	"errors"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"time"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func NewRankingLocalCache() *RankingLocalCache {
	return &RankingLocalCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValueOf[time.Time](time.Now()),
		expiration: time.Minute * 3,
	}
}

func (r *RankingLocalCache) Set(_ context.Context, arts []domain.Article) error {
	r.ddl.Store(time.Now().Add(time.Minute * 3))
	r.topN.Store(arts)
	return nil
}

func (r *RankingLocalCache) Get(_ context.Context) ([]domain.Article, error) {
	arts := r.topN.Load()
	if len(arts) == 0 || r.ddl.Load().Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (r *RankingLocalCache) ForceGet(_ context.Context) ([]domain.Article, error) {
	return r.topN.Load(), nil
}

package service

import (
	"context"
	"github.com/ac-zht/gotools/queue"
	intrv1 "github.com/ac-zht/webook/interactive/service"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

//go:generate mockgen -source=./ranking.go -package=svcmocks -destination=mocks/ranking.mocks.go RankingService
type RankingService interface {
	RankTopN(ctx context.Context) error
	TopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	intrSvc   intrv1.InteractiveService
	artSvc    ArticleService
	repo      repository.RankingRepository
	BatchSize int
	N         int
	scoreFunc func(likeCnt int64, utime time.Time) float64
}

func NewBatchRankingService(
	intrSvc intrv1.InteractiveService,
	artSvc ArticleService,
	repo repository.RankingRepository) RankingService {
	res := &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		repo:      repo,
		BatchSize: 100,
		N:         100,
	}
	res.scoreFunc = res.score
	return res
}

func (b *BatchRankingService) RankTopN(ctx context.Context) error {
	arts, err := b.rankTopN(ctx)
	if err != nil {
		return err
	}
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) rankTopN(ctx context.Context) ([]domain.Article, error) {
	now := time.Now()
	ddl := now.Add(-time.Hour * 24 * 7)
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	que := queue.NewPriorityQueue[Score](b.N,
		func(src Score, dst Score) int {
			if src.score < dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			}
			return -1
		})
	for {
		arts, err := b.artSvc.ListPub(ctx, now, offset, b.BatchSize)
		if err != nil {
			return nil, err
		}
		artIds := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		intrs, err := b.intrSvc.GetByIds(ctx, "article", artIds) //biz_id=>Interactive
		if err != nil {
			return nil, err
		}
		for _, art := range arts {
			intr, ok := intrs[art.Id]
			if !ok {
				continue
			}
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			val, _ := que.Peek()
			//要和堆顶作比较
			if score > val.score {
				ele := Score{art: art, score: score}
				err = que.Enqueue(ele)
				if err == queue.ErrOutOfCapacity {
					_, _ = que.Dequeue()
					err = que.Enqueue(ele)
				}
			}
		}
		if len(arts) == 0 || len(arts) < b.BatchSize ||
			arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
		offset = offset + len(arts)
	}
	ql := que.Len()
	res := make([]domain.Article, ql)
	for i := ql - 1; i >= 0; i-- {
		val, _ := que.Dequeue()
		res[i] = val.art
	}
	return res, nil
}

func (b *BatchRankingService) TopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

func (b *BatchRankingService) score(likeCnt int64, utime time.Time) float64 {
	const factor = 1.5
	return float64(likeCnt-1) /
		math.Pow(time.Since(utime).Hours()+2, factor)
}

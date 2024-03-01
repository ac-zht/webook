package service

import (
	"context"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository"
	"github.com/zht-account/webook/pkg/logger"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	List(ctx context.Context, author int64, offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)

	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, startTime time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	authorRepo repository.ArticleAuthorRepository
	readerRepo repository.ArticleReaderRepository

	repo   repository.ArticleRepository
	logger logger.Logger
}

func NewArticleService(repo repository.ArticleRepository, l logger.Logger) ArticleService {
	return &articleService{
		repo:   repo,
		logger: l,
	}
}

func NewArticleServiceV1(
	authorRepo repository.ArticleAuthorRepository,
	readerRepo repository.ArticleReaderRepository,
	l logger.Logger) ArticleService {
	return &articleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
		logger:     l,
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.update(ctx, art)
		return art.Id, err
	}
	return a.create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id == 0 {
		id, err = a.authorRepo.Create(ctx, art)
	} else {
		err = a.authorRepo.Update(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, art)
		if err == nil {
			break
		}
		a.logger.Error("部分失败： 保存数据到线上库失败",
			logger.Field{Key: "art_id", Value: id},
			logger.Error(err))
	}
	if err != nil {
		a.logger.Error("部分失败：保存数据到线上库重试都失败",
			logger.Field{Key: "art_id", Value: id},
			logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (a *articleService) Withdraw(ctx context.Context, uid, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (a *articleService) List(ctx context.Context, author int64, offset, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, author, offset, limit)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	return a.repo.GetPublishedById(ctx, id)
}

func (a *articleService) ListPub(ctx context.Context, startTime time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, startTime, offset, limit)
}

func (a *articleService) create(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Create(ctx, art)
}

func (a *articleService) update(ctx context.Context, art domain.Article) error {
	return a.repo.Update(ctx, art)
}

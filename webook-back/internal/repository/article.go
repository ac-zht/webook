package repository

import (
	"context"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository/cache"
	"github.com/zht-account/webook/internal/repository/dao/article"
	"github.com/zht-account/webook/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	//List(ctx context.Context, author int64, offset int, limit int) ([]domain.Article, error)

	Sync(ctx context.Context, arr domain.Article) (int64, error)
	//SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error
	//GetById(ctx context.Context, id int64) (domain.Article, error)
}

type ArticleAuthorRepository interface {
}

type ArticleReaderRepository interface {
}

type CachedArticleRepository struct {
	dao article.ArticleDAO

	userRepo UserRepository
	cache    cache.ArticleCache

	//authorDAO article.ArticleAuthorDAO
	//readerDAO article.ArticleReaderDAO

	//db *gorm.DB
	l logger.Logger
}

func NewArticleRepository(dao article.ArticleDAO,
	c cache.ArticleCache,
	userRepo UserRepository,
	l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		l:        l,
		cache:    c,
		userRepo: userRepo,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err != nil {
		return 0, err
	}
	author := art.Author.Id
	err = c.cache.DelFirstPage(ctx, author)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("author", author), logger.Error(err))
	}
	return id, nil
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err != nil {
		return err
	}
	author := art.Author.Id
	err = c.cache.DelFirstPage(ctx, author)
	if err != nil {
		c.l.Error("删除缓存失败", logger.Int64("author", author), logger.Error(err))
	}
	return nil
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err != nil {
		return 0, err
	}
	go func() {
		author := art.Author.Id
		err = c.cache.DelFirstPage(ctx, author)
		if err != nil {
			c.l.Error("删除第一页缓存失败", logger.Int64("author", author), logger.Error(err))
		}
		user, err := c.userRepo.FindById(ctx, author)
		if err != nil {
			c.l.Error("提前设置缓存准备用户信息失败", logger.Int64("uid", author), logger.Error(err))
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		err = c.cache.SetPub(ctx, art)
		if err != nil {
			c.l.Error("提前设置缓存失败", logger.Int64("author", author), logger.Error(err))
		}
	}()
	return id, nil
}

func (c *CachedArticleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   uint8(art.Status),
	}
}

package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository/cache"
	"github.com/zht-account/webook/internal/repository/dao/article"
	"github.com/zht-account/webook/pkg/logger"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	//List(ctx context.Context, author int64, offset int, limit int) ([]domain.Article, error)

	Sync(ctx context.Context, arr domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)

	List(ctx context.Context, author int64, offset, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, utime time.Time, offset, limit int) ([]domain.Article, error)
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

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	cacheArt, err := c.cache.Get(ctx, id)
	if err == nil {
		return cacheArt, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(art), nil
}

func (c *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	user, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res = domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		},
	}
	go func() {
		if err = c.cache.SetPub(ctx, res); err != nil {
			c.l.Error("缓存已发表文章失败", logger.Error(err), logger.Int64("aid", res.Id))
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) List(ctx context.Context, author int64, offset, limit int) ([]domain.Article, error) {
	if offset == 0 && limit == 100 {
		data, err := c.cache.GetFirstPage(ctx, author)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
		if err != cache.ErrKeyNotExist {
			c.l.Error("查询缓存文章失败", logger.Int64("author", author), logger.Error(err))
		}
	}
	arts, err := c.dao.GetByAuthor(ctx, author, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[article.Article, domain.Article](arts,
		func(idx int, src article.Article) domain.Article {
			return c.toDomain(src)
		})
	go func() {
		c.preCache(ctx, res)
	}()
	err = c.cache.SetFirstPage(ctx, author, res)
	if err != nil {
		c.l.Error("刷新第一页文章的缓存失败", logger.Int64("author", author), logger.Error(err))
	}
	return res, nil
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, utime time.Time, offset, limit int) ([]domain.Article, error) {
	val, err := c.dao.ListPubByUtime(ctx, utime, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[article.PublishedArticle, domain.Article](val, func(idx int, src article.PublishedArticle) domain.Article {
		return c.toDomain(article.Article(src))
	}), nil
}

func (c *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		if err := c.cache.Set(ctx, arts[0]); err != nil {
			c.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}

func (c *CachedArticleRepository) toDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
	}
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

package repository

import (
	"context"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository/dao/article"
)

type ArticleReaderRepository interface {
	Save(ctx context.Context, article domain.Article) error
}

type CachedArticleReaderRepository struct {
	dao article.ArticleReaderDAO
}

func (c *CachedArticleReaderRepository) Save(ctx context.Context, article domain.Article) error {
	return c.dao.Upsert(ctx, c.toEntity(article))
}

func NewArticleReaderRepository(dao article.ArticleReaderDAO) ArticleReaderRepository {
	return &CachedArticleReaderRepository{
		dao: dao,
	}
}

func (c *CachedArticleReaderRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}

package repository

import (
	"context"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/repository/dao/article"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

type CachedArticleAuthorRepository struct {
	dao article.ArticleAuthorDAO
}

func (c *CachedArticleAuthorRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Create(ctx, c.toEntity(art))
}

func (c *CachedArticleAuthorRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c *CachedArticleAuthorRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}

func NewArticleAuthorRepository(dao article.ArticleAuthorDAO) ArticleAuthorRepository {
	return &CachedArticleAuthorRepository{
		dao: dao,
	}
}

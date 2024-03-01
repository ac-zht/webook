package article

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleReaderDAO interface {
	// Upsert 不同库同表名
	Upsert(ctx context.Context, art Article) error
	//同库不同表
	UpsertV2(ctx context.Context, art PublishedArticle) error
}

type GORMArticleReaderDAO struct {
	db *gorm.DB
}

func (dao *GORMArticleReaderDAO) Upsert(ctx context.Context, art Article) error {
	return dao.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
		}),
	}).Create(&art).Error
}

func (dao *GORMArticleReaderDAO) UpsertV2(ctx context.Context, art PublishedArticle) error {
	return dao.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
		}),
	}).Create(&art).Error
}

func NewGORMArticleReaderDAO(db *gorm.DB) ArticleReaderDAO {
	return &GORMArticleReaderDAO{
		db: db,
	}
}

package dao

import (
	"github.com/zht-account/webook/interactive/repository/dao"
	"github.com/zht-account/webook/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{},
		&article.Article{},
		&article.PublishedArticle{},
		&dao.Interactive{},
		&dao.UserLikeBiz{},
		&dao.Collection{},
		&dao.UserCollectionBiz{},
	)
}

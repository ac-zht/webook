package article

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"github.com/zht-account/webook/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

var statusPrivate = domain.ArticleStatusPrivate.ToUint8()

type S3DAO struct {
	oss *s3.S3
	GORMArticleDAO
	bucket *string
}

func NewOssDAO(oss *s3.S3, db *gorm.DB) ArticleDAO {
	return &S3DAO{
		oss:    oss,
		bucket: ekit.ToPtr[string]("webook-18627502290"),
		GORMArticleDAO: GORMArticleDAO{
			db: db,
		},
	}
}

func (o *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	err := o.db.Transaction(func(tx *gorm.DB) error {
		var (
			id  = art.Id
			err error
		)
		now := time.Now().UnixMilli()
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishedArt := PublishedArticle{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Status:   art.Status,
			Ctime:    now,
			Utime:    now,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  publishedArt.Title,
				"status": publishedArt.Status,
				"utime":  now,
			}),
		}).Create(&publishedArt).Error
	})
	if err != nil {
		return 0, err
	}
	_, err = o.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      o.bucket,
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return art.Id, err
}

func (o *S3DAO) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id=? AND author_id = ?", id, author).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		res = tx.Model(&PublishedArticle{}).
			Where("id=? AND author_id = ?", id, author).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
	if err != nil {
		return err
	}
	if status == statusPrivate {
		_, err = o.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: o.bucket,
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}
	return err
}

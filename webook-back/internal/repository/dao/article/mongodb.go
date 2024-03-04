package article

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDDAO struct {
	col     *mongo.Collection
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func (m MongoDDAO) Insert(ctx context.Context, art Article) (int64, error) {
	art.Id = m.node.Generate().Int64()
	now := time.Now().UnixMilli()
	art.Utime = now
	art.Ctime = now
	_, err := m.col.InsertOne(ctx, art)
	return art.Id, err
}

func (m MongoDDAO) UpdateById(ctx context.Context, art Article) error {
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.D{bson.E{Key: "title", Value: art.Title},
			bson.E{Key: "content", Value: art.Content},
			bson.E{Key: "status", Value: art.Status},
			bson.E{Key: "utime", Value: time.Now().UnixMilli()},
		}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.MatchedCount != 1 {
		return errors.New("更新失败")
	}
	return nil
}

func (m MongoDDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return id, err
	}
	art.Id = id
	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	now := time.Now().UnixMilli()
	art.Utime = now
	_, err = m.liveCol.UpdateOne(ctx, filter, bson.D{bson.E{Key: "$set", Value: art},
		bson.E{Key: "$setOnInsert", Value: bson.D{bson.E{Key: "ctime", Value: now}}}},
		options.Update().SetUpsert(true))
	return id, err
}

func (m MongoDDAO) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	filter := bson.D{bson.E{Key: "id", Value: id},
		bson.E{Key: "author_id", Value: author}}
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.D{bson.E{Key: "status", Value: status}}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func (m MongoDDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m MongoDDAO) ListPubByUtime(ctx context.Context, utime time.Time, offset, limit int) ([]PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().CreateMany(ctx, index)
	return err
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDDAO{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

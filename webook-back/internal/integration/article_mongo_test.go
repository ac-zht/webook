package integration

import (
    "context"
    "github.com/bwmarrin/snowflake"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/zht-account/webook/internal/domain"
    "github.com/zht-account/webook/internal/integration/startup"
    "github.com/zht-account/webook/internal/repository/dao/article"
    ijwt "github.com/zht-account/webook/internal/web/jwt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "testing"
    "time"
)

type ArticleMongoHandlerTestSuite struct {
    suite.Suite
    server  *gin.Engine
    mdb     *mongo.Database
    col     *mongo.Collection
    liveCol *mongo.Collection
}

func (s *ArticleMongoHandlerTestSuite) SetupSuite() {
    s.server = gin.Default()
    s.server.Use(func(context *gin.Context) {
        context.Set("user", ijwt.UserClaims{
            Id: 3,
        })
        context.Next()
    })
    s.mdb = startup.InitMongoDB()
    node, err := snowflake.NewNode(1)
    assert.NoError(s.T(), err)
    err = article.InitCollections(s.mdb)
    if err != nil {
        panic(err)
    }
    s.col = s.mdb.Collection("articles")
    s.liveCol = s.mdb.Collection("published_articles")
    hdl := startup.InitArticleHandler(article.NewMongoDBDAO(s.mdb, node))
    hdl.RegisterRoutes(s.server)
}

func (s *ArticleMongoHandlerTestSuite) TearDownTest() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
    defer cancel()
    _, err := s.mdb.Collection("articles").
        DeleteMany(ctx, bson.D{})
    assert.NoError(s.T(), err)
    _, err = s.mdb.Collection("published_articles").
        DeleteMany(ctx, bson.D{})
    assert.NoError(s.T(), err)
}

func (s *ArticleMongoHandlerTestSuite) TestArticleHandler_Edit() {
    t := s.T()
    testCase := []struct {
        name       string
        before     func(t *testing.T)
        after      func(t *testing.T)
        req        Article
        wantCode   int
        wantResult Result[int64]
    }{
        {
            name: "新建帖子",
            before: func(t *testing.T) {

            },
            after: func(t *testing.T) {
                ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
                defer cancel()
                var art article.Article
                err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 3}}).Decode(&art)
                assert.NoError(t, err)
                assert.True(t, art.Ctime > 0)
                assert.True(t, art.Utime > 0)
                assert.True(t, art.Id > 0)
                art.Utime = 0
                art.Ctime = 0
                art.Id = 0
                assert.Equal(t, article.Article{
                    Title:    "hello，你好",
                    Content:  "随便试试",
                    AuthorId: 3,
                    Status:   domain.ArticleStatusUnpublished.ToUint8(),
                }, art)
            },
            req: Article{
                Title:   "hello，你好",
                Content: "随便试试",
            },
            wantCode: 200,
            wantResult: Result[int64]{
                Data: 1,
            },
        },
        {
            name: "更新帖子",
            before: func(t *testing.T) {
                ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
                defer cancel()
                _, err := s.col.InsertOne(ctx, &article.Article{
                    Id:       2,
                    Title:    "我的标题",
                    Content:  "我的内容",
                    Ctime:    456,
                    Utime:    234,
                    AuthorId: 3,
                    Status:   domain.ArticleStatusPublished.ToUint8(),
                })
                assert.NoError(t, err)
            },
            after: func(t *testing.T) {
                ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
                defer cancel()
                var art article.Article
                err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
                assert.NoError(t, err)
                assert.True(t, art.Utime > 234)
                art.Utime = 0
                assert.Equal(t, article.Article{
                    Id:       2,
                    Title:    "新的标题",
                    Content:  "新的内容",
                    AuthorId: 3,
                    Ctime:    456,
                    Status:   domain.ArticleStatusUnpublished.ToUint8(),
                }, art)
            },
            req: Article{
                Id:      2,
                Title:   "新的标题",
                Content: "新的内容",
            },
            wantCode: 200,
            wantResult: Result[int64]{
                Data: 2,
            },
        },
        {
            name: "更新别人的帖子",
            before: func(t *testing.T) {
                ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
                defer cancel()
                _, err := s.col.InsertOne(ctx, &article.Article{
                    Id:       3,
                    Title:    "我的标题",
                    Content:  "我的内容",
                    Ctime:    456,
                    Utime:    234,
                    AuthorId: 4,
                    Status:   domain.ArticleStatusUnpublished.ToUint8(),
                })
                assert.NoError(t, err)
            },
            after: func(t *testing.T) {
                ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
                defer cancel()
                var art article.Article
                err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
                assert.NoError(t, err)
                assert.True(t, art.Utime > 234)
                art.Utime = 0
                assert.Equal(t, article.Article{
                    Id:       2,
                    Title:    "新的标题",
                    Content:  "新的内容",
                    AuthorId: 3,
                    Ctime:    456,
                    Status:   domain.ArticleStatusUnpublished.ToUint8(),
                }, art)
            },
            req: Article{
                Id:      2,
                Title:   "新的标题",
                Content: "新的内容",
            },
            wantCode: 200,
            wantResult: Result[int64]{
                Data: 2,
            },
        },
    }
}

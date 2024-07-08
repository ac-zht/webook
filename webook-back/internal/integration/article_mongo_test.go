package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/errs"
	"github.com/ac-zht/webook/internal/integration/startup"
	"github.com/ac-zht/webook/internal/repository/dao/article"
	ijwt "github.com/ac-zht/webook/internal/web/jwt"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
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

func TestMongoArticle(t *testing.T) {
	suite.Run(t, new(ArticleMongoHandlerTestSuite))
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
	testCases := []struct {
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
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 4,
					Ctime:    456,
					Utime:    234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: errs.ArticleInternalServerError,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, result.Code)
			// 只能判定有 ID，因为雪花算法你无法确定具体的值
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

func (s *ArticleMongoHandlerTestSuite) TestArticle_Publish() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		req    Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "hello，你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(3), art.AuthorId)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 3}}).Decode(&publishedArt)
				assert.Equal(t, "hello，你好", publishedArt.Title)
				assert.Equal(t, "随便试试", publishedArt.Content)
				assert.Equal(t, int64(3), publishedArt.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
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
			name: "更新帖子并新发表",
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
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, int64(2), art.Id)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(3), art.AuthorId)
				assert.Equal(t, int64(456), art.Ctime)
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, int64(2), art.Id)
				assert.Equal(t, "新的标题", publishedArt.Title)
				assert.Equal(t, "新的内容", publishedArt.Content)
				assert.Equal(t, int64(3), publishedArt.AuthorId)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
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
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 3,
				}
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, int64(3), art.Id)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(3), art.AuthorId)
				assert.Equal(t, int64(456), art.Ctime)
				assert.True(t, art.Utime > 234)

				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, int64(3), publishedArt.Id)
				assert.Equal(t, "新的标题", publishedArt.Title)
				assert.Equal(t, "新的内容", publishedArt.Content)
				assert.Equal(t, int64(3), publishedArt.AuthorId)
				assert.Equal(t, int64(456), publishedArt.Ctime)
				assert.True(t, publishedArt.Utime > 234)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				art := article.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
				}
				// 模拟已经存在的帖子，并且是已经发布的帖子
				_, err := s.col.InsertOne(ctx, &art)
				assert.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				// 验证一下数据
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, int64(4), art.Id)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, int64(789), art.AuthorId)

				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&publishedArt)
				assert.NoError(t, err)
				assert.Equal(t, int64(4), publishedArt.Id)
				assert.Equal(t, "我的标题", publishedArt.Title)
				assert.Equal(t, "我的内容", publishedArt.Content)
				assert.Equal(t, int64(789), publishedArt.AuthorId)
				assert.Equal(t, int64(456), publishedArt.Ctime)
				assert.Equal(t, int64(234), publishedArt.Utime)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: errs.ArticleInternalServerError,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult.Code, result.Code)
			if tc.wantResult.Data > 0 {
				assert.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

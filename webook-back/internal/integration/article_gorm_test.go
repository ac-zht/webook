package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	intr "github.com/ac-zht/webook/interactive/repository/dao"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/errs"
	articleEvents "github.com/ac-zht/webook/internal/events/article"
	"github.com/ac-zht/webook/internal/integration/startup"
	"github.com/ac-zht/webook/internal/repository/dao/article"
	ijwt "github.com/ac-zht/webook/internal/web/jwt"
	"github.com/ac-zht/webook/pkg/saramax"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const topicReadEvent = "article_read_event"

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArticleGORMHandlerTestSuite struct {
	suite.Suite
	server      *gin.Engine
	db          *gorm.DB
	kafkaClient sarama.Client
}

func TestGORMArticle(t *testing.T) {
	suite.Run(t, new(ArticleGORMHandlerTestSuite))
}

func (a *ArticleGORMHandlerTestSuite) SetupSuite() {
	a.server = gin.Default()
	a.server.Use(func(context *gin.Context) {
		//设置好当前用户ID
		context.Set("user", ijwt.UserClaims{
			Id: 3,
		})
		context.Next()
	})
	a.db = startup.InitTestDB()
	a.kafkaClient = startup.InitKafka()
	hdl := startup.InitArticleHandler(article.NewGORMArticleDAO(a.db))
	hdl.RegisterRoutes(a.server)
}

func (a *ArticleGORMHandlerTestSuite) SetupTest() {
	err := a.db.Exec("TRUNCATE TABLE `articles`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("TRUNCATE TABLE `published_articles`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(a.T(), err)
}

func (a *ArticleGORMHandlerTestSuite) TestArticleHandler_Edit() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		req    Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art article.Article
				a.db.Where("author_id = ?", 3).First(&art)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, article.Article{
					Id:       1,
					Title:    "hello, 你好",
					Content:  "随便试试",
					AuthorId: 3,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			req: Article{
				Title:   "hello, 你好",
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
				a.db.Create(&article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 3,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
			},
			after: func(t *testing.T) {
				var art article.Article
				a.db.Where("id = ?", 2).First(&art)
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
				// 模拟已经存在的帖子
				a.db.Create(&article.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				var art article.Article
				a.db.Where("id = ?", 3).First(&art)
				assert.Equal(t, article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
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
		a.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

func (a *ArticleGORMHandlerTestSuite) TestArticle_Publish() {
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
				var art article.Article
				a.db.Where("author_id = ?", 3).First(&art)
				assert.Equal(t, "hello，你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(3), art.AuthorId)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				var publishedArt article.PublishedArticle
				a.db.Where("author_id = ?", 3).First(&publishedArt)
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
				a.db.Create(&article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 3,
				})
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art article.Article
				a.db.Where("id = ?", 2).First(&art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(3), art.AuthorId)
				assert.Equal(t, int64(456), art.Ctime)
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				a.db.Where("id = ?", 2).First(&publishedArt)
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
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 3,
				}
				a.db.Create(&art)
				part := article.PublishedArticle(art)
				a.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art article.Article
				a.db.Where("id = ?", 3).First(&art)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(3), art.AuthorId)
				assert.Equal(t, int64(456), art.Ctime)
				assert.True(t, art.Utime > 234)

				var publishedArt article.PublishedArticle
				a.db.Where("id = ?", 3).First(&publishedArt)
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
				art := article.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
				}
				a.db.Create(&art)
				part := article.PublishedArticle(article.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
				})
				a.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art article.Article
				a.db.Where("id = ?", 4).First(&art)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, int64(789), art.AuthorId)

				var publishedArt article.PublishedArticle
				a.db.Where("id = ?", 4).First(&publishedArt)
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
		a.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()
			a.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

func (a *ArticleGORMHandlerTestSuite) TestPubDetail() {
	testCases := []struct {
		name       string
		before     func(t *testing.T)
		after      func(t *testing.T)
		id         int64
		wantCode   int
		wantResult Result[Article]
	}{
		{
			name: "查找成功",
			id:   1,
			before: func(t *testing.T) {
				err := a.db.Create(&article.PublishedArticle{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 3,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    123,
					Utime:    234,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&intr.Interactive{
					Id:         1,
					BizId:      3,
					Biz:        "article",
					ReadCnt:    1,
					CollectCnt: 2,
					LikeCnt:    3,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				consumer, err := sarama.NewConsumerGroupFromClient("test_group1",
					a.kafkaClient)
				assert.NoError(t, err)
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				err = consumer.Consume(ctx, []string{topicReadEvent}, saramax.HandlerFunc(func(session sarama.ConsumerGroupSession,
					claim sarama.ConsumerGroupClaim) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case msg := <-claim.Messages():
						var evt articleEvents.ReadEvent
						err := json.Unmarshal(msg.Value, &evt)
						if err != nil {
							return err
						}
						assert.Equal(t, articleEvents.ReadEvent{
							Aid: 1,
							Uid: 3,
						}, evt)
					}
					return nil
				}))
				assert.NoError(t, err)
			},
			wantCode: 200,
			wantResult: Result[Article]{
				Data: Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
				},
			},
		},
	}

	for _, tc := range testCases {
		a.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodGet,
				fmt.Sprintf("/articles/pub/%d", tc.id), nil)
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			var result Result[Article]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

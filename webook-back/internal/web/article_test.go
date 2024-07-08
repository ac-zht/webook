package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ac-zht/webook/internal/domain"
	"github.com/ac-zht/webook/internal/errs"
	"github.com/ac-zht/webook/internal/service"
	svcmocks "github.com/ac-zht/webook/internal/service/mocks"
	"github.com/ac-zht/webook/internal/web/jwt"
	"github.com/ac-zht/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) service.ArticleService
		reqBody string

		wantCode int
		wantRes  Result
	}{
		{
			name: "新建立刻发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 789,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
    {
       "title":"我的标题",
       "content":"我的内容"
    }`,
			wantCode: 200,
			wantRes: Result{
				Msg:  "OK",
				Data: float64(1),
			},
		},
		{
			name: "已有帖子发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      12,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 789,
					},
				}).Return(int64(12), nil)
				return svc
			},
			reqBody: `
    {
       "id": 12,
       "title": "我的标题",
       "content":"我的内容"
    }`,
			wantCode: 200,
			wantRes: Result{
				Msg:  "OK",
				Data: float64(12),
			},
		},
		{
			name: "发表失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 789,
					},
				}).Return(int64(0), errors.New("mock 错误"))
				return svc
			},
			reqBody: `
    {
       "title": "我的标题",
       "content":"我的内容"
    }`,
			wantCode: 200,
			wantRes: Result{
				Code: errs.ArticleInternalServerError,
				Msg:  "系统错误",
			},
		},
		{
			name: "Bind 错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
    {
       "title": "我的标题",
       "cont
    }`,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		fmt.Println(tc.name)
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := tc.mock(ctrl)
			hdl := NewArticleHandler(svc, nil, logger.NewNoOpLogger())
			gin.SetMode(gin.TestMode)

			//注册路由
			server := gin.Default()
			//设置登录态
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", jwt.UserClaims{
					Id: 789,
				})
			})
			hdl.RegisterRoutes(server)
			//准备请求
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish",
				bytes.NewReader([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			//准备记录响应
			recorder := httptest.NewRecorder()
			//执行
			server.ServeHTTP(recorder, req)
			//断言
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var res Result
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

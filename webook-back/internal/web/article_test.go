package web

import (
	"github.com/zht-account/webook/internal/domain"
	"go.uber.org/mock/gomock"
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
			name: "新建l发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			reqBody: `
{
    "title":"我的标题",
    "content":"我的内容"
}`,
			wantCode: 200,
			wantRes:  Result{
				//Data: float64(1),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
		})
	}
}

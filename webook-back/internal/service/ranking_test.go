package service

import (
	domain2 "github.com/zht-account/webook/interactive/domain"
	"github.com/zht-account/webook/interactive/service"
	"github.com/zht-account/webook/internal/domain"
	svcmocks "github.com/zht-account/webook/internal/service/mocks"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestBatchRankingService_rankTopN(t *testing.T) {
	const batchSize = 2
	now := time.Now()
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (service.InteractiveService, ArticleService)
		wantErr error
		wantRes []domain.Article
	}{
		{
			name: "计算成功-两批次",
			mock: func(ctrl *gomock.Controller) (service.InteractiveService, ArticleService) {
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).
					Return([]domain.Article{
						{Id: 1, Utime: now},
						{Id: 2, Utime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).
					Return([]domain.Article{
						{Id: 4, Utime: now},
						{Id: 3, Utime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, batchSize).
					Return([]domain.Article{}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2}).
					Return(map[int64]domain2.Interactive{
						1: {LikeCnt: 1},
						2: {LikeCnt: 2},
					}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{4, 3}).
					Return(map[int64]domain2.Interactive{
						3: {LikeCnt: 3},
						4: {LikeCnt: 4},
					}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).
					Return(map[int64]domain2.Interactive{}, nil)
				return intrSvc, artSvc
			},
			wantRes: []domain.Article{
				{Id: 4, Utime: now},
				{Id: 3, Utime: now},
				{Id: 2, Utime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			//intrSvc, artSvc := tc.mock(ctrl)
		})
	}
}

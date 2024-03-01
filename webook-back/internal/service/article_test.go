package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/repository"
	articlerepomocks "github.com/zht-account/webook/internal/repository/mocks"
	"github.com/zht-account/webook/pkg/logger"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestArticleHandler_Edit(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (
			repository.ArticleAuthorRepository,
			repository.ArticleReaderRepository)

		art domain.Article

		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				ar := articlerepomocks.NewMockArticleAuthorRepository(ctrl)
				ar.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)

				rr := articlerepomocks.NewMockArticleReaderRepository(ctrl)
				rr.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				return ar, rr
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
		{
			name: "修改保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				ar := articlerepomocks.NewMockArticleAuthorRepository(ctrl)
				ar.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      7,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("保存失败"))

				rr := articlerepomocks.NewMockArticleReaderRepository(ctrl)
				return ar, rr
			},
			art: domain.Article{
				Id:      7,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: errors.New("保存失败"),
		},
		{
			name: "修改保存到线上库失败-重试都失败",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				ar := articlerepomocks.NewMockArticleAuthorRepository(ctrl)
				ar.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      7,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)

				rr := articlerepomocks.NewMockArticleReaderRepository(ctrl)
				rr.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      7,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).AnyTimes().Return(errors.New("保存到线上库失败"))
				return ar, rr
			},
			art: domain.Article{
				Id:      7,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantErr: errors.New("保存到线上库失败"),
		},
		{
			name: "修改保存到线上库失败-重试成功",
			mock: func(ctrl *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				ar := articlerepomocks.NewMockArticleAuthorRepository(ctrl)
				ar.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      7,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)

				rr := articlerepomocks.NewMockArticleReaderRepository(ctrl)
				rr.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      7,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(errors.New("保存到线上失败"))
				rr.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      7,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				return ar, rr
			},
			art: domain.Article{
				Id:      7,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 7,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authorRepo, readerRepo := tc.mock(ctrl)
			svc := NewArticleServiceV1(authorRepo, readerRepo, logger.NewNoOpLogger())
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}

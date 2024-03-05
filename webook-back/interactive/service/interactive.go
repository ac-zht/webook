package service

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/interactive/repository"
	"github.com/zht-account/webook/interactive/domain"
	"github.com/zht-account/webook/pkg/logger"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId, uid int64) error
	CancelLike(ctx context.Context, biz string, bizId, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func (i interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i interactiveService) Like(ctx context.Context, biz string, bizId, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (i interactiveService) CancelLike(ctx context.Context, biz string, bizId, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (i interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (i interactiveService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (i interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}

package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/interactive/repository/cache"
	"gitee.com/geekbang/basic-go/webook/interactive/repository/dao"
	"github.com/zht-account/webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
}

type CachedReadCntRepository struct {
	cache cache.InteractiveCache
	dao   dao.InteractiveDAO
	l     logger.Logger
}

func (c CachedReadCntRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

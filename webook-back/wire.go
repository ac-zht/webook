//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/zht-account/webook/internal/repository"
	"github.com/zht-account/webook/internal/repository/cache"
	"github.com/zht-account/webook/internal/repository/dao"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/web"
	"github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,

		dao.NewUserDAO,

		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,

		repository.NewCachedUserRepository,
		repository.NewCacheCodeRepository,

		service.NewUserService,
		service.NewSMSCodeService,
		ioc.InitSMSService,

		ioc.InitLogger,

		jwt.NewRedisHandler,
		web.NewUserHandler,

		ioc.InitGin,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}

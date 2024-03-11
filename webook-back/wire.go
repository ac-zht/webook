//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	repository2 "github.com/zht-account/webook/interactive/repository"
	cache2 "github.com/zht-account/webook/interactive/repository/cache"
	dao2 "github.com/zht-account/webook/interactive/repository/dao"
	service2 "github.com/zht-account/webook/interactive/service"
	"github.com/zht-account/webook/internal/repository"
	"github.com/zht-account/webook/internal/repository/cache"
	"github.com/zht-account/webook/internal/repository/dao"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/web"
	"github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/ioc"
)

var interactiveServiceProducer = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitApp() *gin.Engine {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewSyncProducer,

		//DAO部分
		dao.NewUserDAO,

		interactiveServiceProducer,

		//cache部分
		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,

		//repository部分
		repository.NewCachedUserRepository,
		repository.NewCacheCodeRepository,

		//service部分
		service.NewUserService,
		service.NewSMSCodeService,
		ioc.InitSMSService,

		//handler部分
		jwt.NewRedisHandler,
		web.NewUserHandler,

		//gin中间件
		ioc.InitMiddlewares,

		//web服务器
		ioc.InitWebServer,
		ioc.InitGin,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}

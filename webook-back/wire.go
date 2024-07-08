//go:build wireinject

package main

import (
	repository2 "github.com/ac-zht/webook/interactive/repository"
	cache2 "github.com/ac-zht/webook/interactive/repository/cache"
	dao2 "github.com/ac-zht/webook/interactive/repository/dao"
	service2 "github.com/ac-zht/webook/interactive/service"
	article2 "github.com/ac-zht/webook/internal/events/article"
	"github.com/ac-zht/webook/internal/repository"
	"github.com/ac-zht/webook/internal/repository/cache"
	"github.com/ac-zht/webook/internal/repository/dao"
	"github.com/ac-zht/webook/internal/repository/dao/article"
	"github.com/ac-zht/webook/internal/service"
	"github.com/ac-zht/webook/internal/web"
	"github.com/ac-zht/webook/internal/web/jwt"
	"github.com/ac-zht/webook/ioc"
	"github.com/google/wire"
)

var interactiveServiceProducer = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewSyncProducer,

		//DAO部分
		dao.NewUserDAO,
		article.NewGORMArticleDAO,

		interactiveServiceProducer,

		//cache部分
		cache.NewRedisUserCache,
		cache.NewRedisCodeCache,
		cache.NewRedisArticleCache,

		//repository部分
		repository.NewCachedUserRepository,
		repository.NewCacheCodeRepository,
		repository.NewArticleRepository,

		//events部分
		article2.NewSaramaSyncProducer,
		ioc.NewConsumers,

		//service部分
		service.NewUserService,
		service.NewSMSCodeService,
		service.NewArticleService,
		ioc.InitSMSService,

		//handler部分
		jwt.NewRedisHandler,
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewObservabilityHandler,

		//gin中间件
		ioc.InitMiddlewares,

		//web服务器
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}

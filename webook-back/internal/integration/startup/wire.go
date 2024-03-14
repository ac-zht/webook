//go:build wireinject

package startup

import (
	"github.com/google/wire"
	repository2 "github.com/zht-account/webook/interactive/repository"
	cache2 "github.com/zht-account/webook/interactive/repository/cache"
	dao2 "github.com/zht-account/webook/interactive/repository/dao"
	service2 "github.com/zht-account/webook/interactive/service"
	article2 "github.com/zht-account/webook/internal/events/article"
	"github.com/zht-account/webook/internal/repository"
	"github.com/zht-account/webook/internal/repository/cache"
	"github.com/zht-account/webook/internal/repository/dao"
	"github.com/zht-account/webook/internal/repository/dao/article"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/web"
)

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB,
	InitLog,
	NewSyncProducer,
	InitKafka,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewRedisUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService)
var articleSvcProvider = wire.NewSet(
	article.NewGORMArticleDAO,
	article2.NewSaramaSyncProducer,
	cache.NewRedisArticleCache,
	repository.NewArticleRepository,
	service.NewArticleService,
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
)

//func InitWebServer() *gin.Engine {
//    wire.Build(
//        thirdProvider,
//        userSvcProvider,
//        articleSvcProvider,
//        interactiveSvcProvider,
//    )
//    return gin.Default()
//}

func InitArticleHandler(dao article.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		interactiveSvcProvider,
		article2.NewSaramaSyncProducer,
		cache.NewRedisArticleCache,
		repository.NewArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler)
	return new(web.ArticleHandler)
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserService(nil, nil)
}

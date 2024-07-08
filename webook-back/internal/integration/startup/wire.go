//go:build wireinject

package startup

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
	"github.com/google/wire"
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

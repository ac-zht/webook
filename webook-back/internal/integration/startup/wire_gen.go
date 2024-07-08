// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitArticleHandler(dao3 article.ArticleDAO) *web.ArticleHandler {
	cmdable := InitRedis()
	articleCache := cache.NewRedisArticleCache(cmdable)
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	logger := InitLog()
	articleRepository := repository.NewArticleRepository(dao3, articleCache, userRepository, logger)
	client := InitKafka()
	syncProducer := NewSyncProducer(client)
	producer := article2.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, logger, producer)
	interactiveDAO := dao2.NewGORMInteractiveDAO(gormDB)
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository2.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, logger)
	interactiveService := service2.NewInteractiveService(interactiveRepository, logger)
	articleHandler := web.NewArticleHandler(articleService, interactiveService, logger)
	return articleHandler
}

func InitUserSvc() service.UserService {
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	logger := InitLog()
	userService := service.NewUserService(userRepository, logger)
	return userService
}

// wire.go:

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB,
	InitLog,
	NewSyncProducer,
	InitKafka,
)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewRedisUserCache, repository.NewCachedUserRepository, service.NewUserService)

var articleSvcProvider = wire.NewSet(article.NewGORMArticleDAO, article2.NewSaramaSyncProducer, cache.NewRedisArticleCache, repository.NewArticleRepository, service.NewArticleService)

var interactiveSvcProvider = wire.NewSet(service2.NewInteractiveService, repository2.NewCachedInteractiveRepository, dao2.NewGORMInteractiveDAO, cache2.NewRedisInteractiveCache)

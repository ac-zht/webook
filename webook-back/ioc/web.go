package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zht-account/webook/internal/web"
	ijwt "github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/internal/web/middleware"
	"github.com/zht-account/webook/pkg/ginx"
	"github.com/zht-account/webook/pkg/ginx/middleware/ratelimit"
	"github.com/zht-account/webook/pkg/logger"
	"strings"
	"time"
)

func InitWebServer(mids []gin.HandlerFunc,
	userHdl *web.UserHandler,
	artHdl *web.ArticleHandler,
	l logger.Logger) *gin.Engine {
	ginx.SetLogger(l)
	server := gin.Default()
	server.Use(mids...)
	userHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, hdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		loginHdl(hdl),
		rateLimitHdl(redisClient),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Jwt-Token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "xxx.com")
		},
		MaxAge: 12 * time.Hour,
	})
}

func loginHdl(hdl ijwt.Handler) gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder(hdl).Build()
}

func rateLimitHdl(redisClient redis.Cmdable) gin.HandlerFunc {
	return ratelimit.NewBuilder(redisClient, time.Minute, 100).Build()
}

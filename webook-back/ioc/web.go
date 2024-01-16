package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zht-account/webook/internal/web"
	"github.com/zht-account/webook/internal/web/middleware"
	"github.com/zht-account/webook/pkg/ginx/middleware/ratelimit"
	"strings"
	"time"
)

func InitGin(mids []gin.HandlerFunc, hdl *web.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mids...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		loginHdl(),
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

func loginHdl() gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder().Build()
}

func rateLimitHdl(redisClient redis.Cmdable) gin.HandlerFunc {
	return ratelimit.NewBuilder(redisClient, time.Minute, 100).Build()
}

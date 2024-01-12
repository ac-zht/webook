package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/zht-account/webook/internal/repository"
	"github.com/zht-account/webook/internal/repository/cache"
	"github.com/zht-account/webook/internal/repository/dao"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/service/sms/tencent"
	"github.com/zht-account/webook/internal/web"
	"github.com/zht-account/webook/internal/web/middleware"
	"github.com/zht-account/webook/pkg/ginx/middleware/ratelimit"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strings"
	"time"
)

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
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
	}))
	//session存储到cookie
	store := cookie.NewStore([]byte("secret"))
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//    []byte("abc"),
	//    []byte("xyz"))
	//if err != nil {
	//    panic(err)
	//}
	server.Use(sessions.Sessions("ssid", store))

	//登录检测
	//login := &middleware.LoginMiddlewareBuilder{}
	login := &middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())

	//请求限流
	cmd := redis.NewClient(&redis.Options{
		Addr:     "120.24.91.113:7002",
		Password: "uphill",
		DB:       2,
	})
	server.Use(ratelimit.NewBuilder(cmd, time.Minute, 100).Build())
	return server
}

func initUser(server *gin.Engine, db *gorm.DB) {
	ud := dao.NewUserDAO(db)
	redisUserCache := cache.NewRedisUserCache(redis.NewClient(&redis.Options{
		Addr:     "120.24.91.113:7002",
		Password: "uphill",
		DB:       2,
	}))
	ur := repository.NewCachedUserRepository(ud, redisUserCache)
	us := service.NewUserService(ur)

	redisCodeCache := cache.NewRedisCodeCache(redis.NewClient(&redis.Options{
		Addr:     "120.24.91.113:7002",
		Password: "uphill",
		DB:       2,
	}))
	codeRepo := repository.NewCacheCodeRepository(redisCodeCache)
	credential := common.NewCredential(
		os.Getenv("TENCENTCLOUD_SECRET_ID"),
		os.Getenv("TENCENTCLOUD_SECRET_KEY"),
	)
	smsClient, _ := sms.NewClient(credential, "ap-guangzhou", profile.NewClientProfile())
	codeSvc := service.NewSMSCodeService(tencent.NewService(smsClient, "", ""), codeRepo)
	c := web.NewUserHandler(us, codeSvc)
	c.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:18627502290@tcp(localhost:3306)/test"))
	if err != nil {
		panic(err)
	}
	//err = dao.InitTables(db)
	//if err != nil {
	//    panic(err)
	//}
	return db
}

func main() {
	//server := gin.Default()
	//server.GET("/hello", func(ctx *gin.Context) {
	//    ctx.String(http.StatusOK, "hello, world")
	//})

	db := initDB()
	server := initWebServer()
	initUser(server, db)

	server.Run(":8080")
}

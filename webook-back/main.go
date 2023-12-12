package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/internal/repository"
	"github.com/zht-account/webook/internal/repository/dao"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/web"
	"github.com/zht-account/webook/internal/web/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "xxx.com")
		},
		MaxAge: 12 * time.Hour,
	}))
	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("ssid", store))
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
	return server
}

func initUser(server *gin.Engine, db *gorm.DB) {
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	c := web.NewUserHandler(us)
	c.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:18627502290@tcp(localhost:3306)/test"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	db := initDB()
	server := initWebServer()
	initUser(server, db)
	server.Run(":8080")
}

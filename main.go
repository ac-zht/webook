package main

import (
    "github.com/gin-gonic/gin"
    "github.com/zht-account/webook/internal/web"
    "net/http"
)

func main() {
    server := gin.Default()
    c := &web.UserHandler{}
    c.RegisterRoutes(server)

    server.GET("/hello", func(ctx *gin.Context) {
        ctx.String(http.StatusOK, "hello, world")
    })

    server.GET("/views/*.html", func(context *gin.Context) {
        path := context.Param(".html")
        context.String(http.StatusOK, "匹配上的值是 %s", path)
    })

    server.GET("/users/:name", func(context *gin.Context) {
        name := context.Param("name")
        context.String(http.StatusOK, "这是你传过来的名字 %s", name)
    })

    server.GET("/order", func(context *gin.Context) {
        id := context.Query("id")
        context.String(http.StatusOK, "你传过来的 ID 是 %s", id)
    })

    server.GET("/views2/*.html", func(context *gin.Context) {
        path := context.Param(".html")
        context.String(http.StatusOK, "匹配上的值是 %s", path)
    })
    server.Run(":8080")
}

package web

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

var _ handler = &UserHandler{}

type UserHandler struct {
}

func (c *UserHandler) Login(ctx *gin.Context) {
}

func (c *UserHandler) SignUp(ctx *gin.Context) {
    ctx.String(http.StatusOK, "正在注册....")
}

func (c *UserHandler) Edit(ctx *gin.Context) {
}

func (c *UserHandler) Profile(ctx *gin.Context) {
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {
    ug := server.Group("/users")
    ug.POST("/signup", c.SignUp)
    ug.POST("/login", c.Login)
    ug.POST("/edit", c.Edit)
    ug.POST("/profile", c.Profile)
}

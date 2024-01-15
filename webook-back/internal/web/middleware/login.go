package middleware

import (
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginMiddlewareBuilder struct {
	publicPaths set.Set[string]
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	s := set.NewMapSet[string](3)
	s.Add("/users/singup")
	s.Add("/users/login_sms/code/send")
	s.Add("/users/login_sms")
	s.Add("/users/login")
	return &LoginMiddlewareBuilder{
		publicPaths: s,
	}
}

func (*LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" ||
			ctx.Request.URL.Path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		if sess.Get("userId") == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

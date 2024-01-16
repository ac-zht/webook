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
	s.Add("/users/signup")
	s.Add("/users/login_sms/code/send")
	s.Add("/users/login_sms")
	s.Add("/users/login")
	return &LoginMiddlewareBuilder{
		publicPaths: s,
	}
}

func (m *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if m.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}
		sess := sessions.Default(ctx)
		if sess.Get("userId") == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

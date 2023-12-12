package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginMiddlewareBuilder struct {
}

func (*LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" ||
			ctx.Request.URL.Path == "/users/login" {
			sess := sessions.Default(ctx)
			if sess.Get("userId") == nil {
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
	}
}

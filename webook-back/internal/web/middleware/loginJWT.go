package middleware

import (
	ijwt "github.com/ac-zht/webook/internal/web/jwt"
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	publicPaths set.Set[string]
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	s := set.NewMapSet[string](3)
	s.Add("/users/signup")
	s.Add("/users/login_sms/code/send")
	s.Add("/users/login_sms")
	s.Add("/users/login")
	s.Add("/test/metric") //prometheus测试链接
	return &LoginJWTMiddlewareBuilder{
		publicPaths: s,
		Handler:     hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if m.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}
		tokenStr := m.ExtractTokenString(ctx)
		if tokenStr == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc := ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.AccessTokenKey, nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if ctx.GetHeader("User-Agent") != uc.UserAgent {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expireTime, err := uc.GetExpirationTime()
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if expireTime.Before(time.Now()) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if m.CheckSession(ctx, uc.Ssid) != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("user", uc)
	}
}

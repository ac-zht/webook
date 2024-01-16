package middleware

import (
	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zht-account/webook/internal/web"
	"log"
	"net/http"
	"strings"
	"time"
)

type LoginJWTMiddlewareBuilder struct {
	publicPaths set.Set[string]
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	s := set.NewMapSet[string](3)
	s.Add("/users/signup")
	s.Add("/users/login_sms/code/send")
	s.Add("/users/login_sms")
	s.Add("/users/login")
	return &LoginJWTMiddlewareBuilder{
		publicPaths: s,
	}
}

func (m *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if m.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}
		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authSegments := strings.SplitN(authCode, " ", 2)
		if len(authSegments) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := authSegments[1]
		uc := web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
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
		ctx.Set("user", uc)
		if expireTime.Sub(time.Now()) < time.Second*50 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
			newToken, err := token.SignedString(web.JWTKey)
			if err != nil {
				log.Println(err)
				return
			}
			ctx.Header("x-jwt-token", newToken)
		}
	}
}

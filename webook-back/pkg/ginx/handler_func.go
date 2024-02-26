package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/pkg/logger"
	"net/http"
)

var log logger.Logger = logger.NewNoOpLogger()

func SetLogger(l logger.Logger) {
	log = l
}

func WrapReq[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}
		res, err := fn(ctx, req)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res.Msg)
	}
}

func WrapJwtReq[Req any](fn func(ctx *gin.Context, req Req, user jwt.UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}
		user, ok := ctx.MustGet("user").(jwt.UserClaims)
		if !ok {
			log.Error("获得用户会话信息失败")
			return
		}
		res, err := fn(ctx, req, user)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res.Msg)
	}
}

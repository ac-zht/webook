package web

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/internal/errs"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/pkg/ginx"
	"github.com/zht-account/webook/pkg/logger"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.Logger
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapJwtReq[ArticleReq](a.Edit))
	g.POST("/publish", ginx.WrapJwtReq[ArticleReq](a.Publish))
	g.POST("/withdraw", ginx.WrapJwtReq[ArticleReq](a.Withdraw))
}

func (a *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, user jwt.UserClaims) (Result, error) {
	id, err := a.svc.Save(ctx, req.toDomain(user.Id))
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{
		Msg:  "OK",
		Data: id,
	}, nil
}

func (a *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, user jwt.UserClaims) (Result, error) {
	id, err := a.svc.Publish(ctx, req.toDomain(user.Id))
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{
		Msg:  "OK",
		Data: id,
	}, err
}

func (a *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleReq, user jwt.UserClaims) (Result, error) {
	if err := a.svc.Withdraw(ctx, user.Id, req.Id); err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, errors.New("设置为尽自己可见失败")
	}
}

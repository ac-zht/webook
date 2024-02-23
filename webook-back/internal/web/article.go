package web

import (
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/internal/service"
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
	g.POST("/edit", ginx.WrapReq[ArticleReq](a.Edit))
}

func (a *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq) (Result, error) {
}

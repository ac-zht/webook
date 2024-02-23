package web

import (
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/pkg/ginx"
)

type handler interface {
	RegisterRoutes(s *gin.Engine)
}

type Result = ginx.Result

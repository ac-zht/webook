package web

import (
	"github.com/ac-zht/webook/pkg/ginx"
	"github.com/gin-gonic/gin"
)

type handler interface {
	RegisterRoutes(s *gin.Engine)
}

type Result = ginx.Result

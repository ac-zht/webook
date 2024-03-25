package web

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type ObservabilityHandler struct {
}

func NewObservabilityHandler() *ObservabilityHandler {
	return &ObservabilityHandler{}
}

func (o *ObservabilityHandler) RegisterRoutes(s *gin.Engine) {
	tg := s.Group("/test")
	tg.GET("/metric", o.Random)
}

func (o *ObservabilityHandler) Random(ctx *gin.Context) {
	num := rand.Int31n(1000)
	time.Sleep(time.Millisecond * time.Duration(num))
	ctx.String(http.StatusOK, "OK")
}

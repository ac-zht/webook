package main

import (
	"github.com/ac-zht/webook/internal/events"
	"github.com/gin-gonic/gin"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}

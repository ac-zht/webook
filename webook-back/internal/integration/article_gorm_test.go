package integration

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/zht-account/webook/internal/integration/startup"
	ijwt "github.com/zht-account/webook/internal/web/jwt"
	"gorm.io/gorm"
)

type ArticleGORMHandlerTestSuite struct {
	suite.Suite
	server      *gin.Engine
	db          *gorm.DB
	kafkaClient sarama.Client
}

func (a *ArticleGORMHandlerTestSuite) SetupSuite() {
	a.server = gin.Default()
	a.server.Use(func(context *gin.Context) {
		context.Set("user", ijwt.UserClaims{
			Id: 123,
		})
		context.Next()
	})
	a.db = startup.InitTestDB()
	a.kafkaClient = startup.InitKafka()
}

func (a *ArticleGORMHandlerTestSuite) TearDownTest() {
	//TODO implement me
	panic("implement me")
}

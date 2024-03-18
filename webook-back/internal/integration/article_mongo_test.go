package integration

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	ijwt "github.com/zht-account/webook/internal/web/jwt"
	"go.mongodb.org/mongo-driver/mongo"
)

type ArticleMongoHandlerTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoHandlerTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(context *gin.Context) {
		context.Set("user", ijwt.UserClaims{
			Id: 3,
		})
		context.Next()
	})
}

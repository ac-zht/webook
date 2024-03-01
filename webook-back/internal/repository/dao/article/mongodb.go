package article

import (
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDDAO struct {
	col     *mongo.Collection
	liveCol *mongo.Collection
	node    *snowflake.Node
}

package startup

import (
	"context"
	"github.com/go-redis/redis/v9"
)

var redisClient redis.Cmdable

func InitRedis() redis.Cmdable {
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "120.24.91.113:7002",
			DB:       2,
			Password: "uphill",
		})
		for err := redisClient.Ping(context.Background()).Err(); err != nil; {
			panic(err)
		}
	}
	return redisClient
}

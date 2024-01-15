package ioc

import "github.com/redis/go-redis/v9"

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:     "120.24.91.113:7002",
		Password: "uphill",
		DB:       2,
	})
}

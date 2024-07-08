package ioc

import (
	"github.com/go-redis/redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
}

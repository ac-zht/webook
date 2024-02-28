package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zht-account/webook/internal/domain"
	"time"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (r *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return r.cmd.Del(ctx, r.Key(id)).Err()
}

func (r *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := r.Key(id)
	data, err := r.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), u)
	return u, err
}

func (r *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := r.Key(u.Id)
	return r.cmd.Set(ctx, key, data, r.expiration).Err()
}

func (r *RedisUserCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 10,
	}
}

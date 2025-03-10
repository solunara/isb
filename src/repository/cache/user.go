package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/model"
)

type UserCache interface {
	Get(ctx context.Context, uid int64) (model.User, error)
	Set(ctx context.Context, u model.User, expiration time.Duration) error
}

type RedisUserCache struct {
	cmd redis.Cmdable
	// expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd: cmd,
	}
}

func (c *RedisUserCache) Set(ctx context.Context, u model.User, expiration time.Duration) error {
	key := c.key(u.Id)
	// JSON序列化一下
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, expiration).Err()
}

func (c *RedisUserCache) Get(ctx context.Context, uid int64) (model.User, error) {
	key := c.key(uid)

	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return model.User{}, err
	}
	var u model.User
	// JSON反序列化一下
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (c *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}

package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaSlideWindow string

// RedisSlidingWindowLimiter Redis 上的滑动窗口算法限流器实现
type RedisSlideWindowLimit struct {
	cmd redis.Cmdable
	// 窗口大小, interval时间内允许rate个请求
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisSlideWindowLimit(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlideWindowLimit{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlideWindowLimit) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key}, r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}

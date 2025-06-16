package cache

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/setCaptcha.lua
var luaSetCaptcha string

//go:embed lua/verifyCaptcha.lua
var luaVerifyCaptcha string

type CaptchaCache interface {
	Set(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error)
}

type RedisCaptchaCache struct {
	cmd redis.Cmdable
}

func NewCaptchaCache(cmd redis.Cmdable) CaptchaCache {
	return &RedisCaptchaCache{
		cmd: cmd,
	}
}

func (c *RedisCaptchaCache) Set(ctx context.Context, biz string, phone string, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCaptcha, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		// 发送太频繁
		return ErrSendTooFrequently
	default:
		// 系统错误
		return ErrSystemError
	}
}

func (c *RedisCaptchaCache) Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCaptcha, []string{c.key(biz, phone)}, inputcode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		// 没有验证次数
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		// 验证码错误
		return false, ErrWrongCode
	}
	return false, ErrUnknownCode
}

func (c *RedisCaptchaCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

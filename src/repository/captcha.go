package repository

import (
	"context"

	"github.com/solunara/isb/src/repository/cache"
)

type CaptchaRepository interface {
	Store(ctx context.Context, biz string, phone string, captcha string) error
	Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error)
}

type captchaRepository struct {
	cache cache.CaptchaCache
}

func NewCaptchaRepository(c cache.CaptchaCache) CaptchaRepository {
	return &captchaRepository{
		cache: c,
	}
}

func (c *captchaRepository) Store(ctx context.Context, biz string, phone string, captcha string) error {
	return c.cache.Set(ctx, biz, phone, captcha)
}

func (c *captchaRepository) Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, inputcode)
}

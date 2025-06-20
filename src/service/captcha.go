package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/service/sms"
)

var (
	ErrSentCaptchaTooOften = errors.New("send too frequently")
)

var captchaTplId = "186224"

type CaptchaService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error)
}

type captchaService struct {
	repo   repository.CaptchaRepository
	smsSvc sms.Service
	tplId  string
}

func NewCaptchaService(repo repository.CaptchaRepository, smsSvc sms.Service, tplId string) CaptchaService {
	return &captchaService{
		repo:   repo,
		smsSvc: smsSvc,
		tplId:  tplId,
	}
}

// biz 用于区别业务场景
func (c *captchaService) Send(ctx context.Context, biz string, phone string) error {
	captcha := c.generateCaptcha()
	err := c.repo.Store(ctx, biz, phone, captcha)
	if err != nil {
		return err
	}
	return c.smsSvc.Send(ctx, c.tplId, []string{captcha}, phone)
}

func (c *captchaService) Verify(ctx context.Context, biz string, phone string, inputcode string) (bool, error) {
	return c.repo.Verify(ctx, biz, phone, inputcode)
}

func (c *captchaService) generateCaptcha() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%6d", num)
}

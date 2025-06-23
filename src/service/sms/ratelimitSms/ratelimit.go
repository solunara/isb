package ratelimitSms

import (
	"context"
	"errors"

	"github.com/solunara/isb/src/pkg/ratelimit"
	"github.com/solunara/isb/src/service/sms"
)

var errLimited = errors.New("触发限流")

var _ sms.Service = &RateLimitSMSService{}

type RateLimitSMSService struct {
	// 被装饰的
	svc     sms.Service
	limiter ratelimit.Limiter
	key     string
}

type RateLimitSMSServiceV1 struct {
	sms.Service
	limiter ratelimit.Limiter
	key     string
}

func (r *RateLimitSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		return errLimited
	}
	return r.svc.Send(ctx, tplId, args, numbers...)
}

func NewRateLimitSMSService(svc sms.Service, l ratelimit.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: l,
		key:     "sms-limiter",
	}
}

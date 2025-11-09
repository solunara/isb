package interceptor

import (
	"context"
	"fmt"
	"strings"

	"github.com/solunara/isb/microgrpc"
	"github.com/solunara/isb/pkg/logger"
	"github.com/solunara/isb/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
	l       logger.Logger
}

func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.Logger) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key, l: l}
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		// 全部限流
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 保守做法
			b.l.Error("判断限流出现错误，", logger.Error(err))
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
			// 激进做法
			// return handler(ctx, req)
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}

		// 针对服务级别的限流
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			limited, err := b.limiter.Limit(ctx, b.key)
			if err != nil {
				// 保守做法
				b.l.Error("判断限流出现错误，", logger.Error(err))
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
				// 激进做法
				// return handler(ctx, req)
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}

		// 针对业务进行限流,比如只限制获取用户id的请求
		if idReq, ok := req.(*microgrpc.GetByIdRequest); ok {
			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("limiter:user:%s:%d", info.FullMethod, idReq.Id))
			if err != nil {
				// 保守做法
				b.l.Error("判断限流出现错误，", logger.Error(err))
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
				// 激进做法
				// return handler(ctx, req)
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}

		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		limted, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 保守做法
			b.l.Error("判断限流出现错误，", logger.Error(err))
			return status.Errorf(codes.ResourceExhausted, "限流")
			// 激进做法
			// return invoker(ctx, method, req, reply, cc, opts...)
		}
		if limted {
			return status.Errorf(codes.ResourceExhausted, "限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// 熔断判断
func (b *InterceptorBuilder) allow() bool {
	// 熔断 的判断条件
	// 常用的是从prometheus里面拿数据来判断
	// prometheus.DefaultGatherer.Gather()
	return false
}

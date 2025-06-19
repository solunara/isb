package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/repository/cache/redismocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) redis.Cmdable
		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(0))
				res.EXPECT().Eval(gomock.Any(), luaSetCaptcha,
					[]string{keyFunc("test", "15212345678")},
					[]any{"123456"}).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: nil,
		},

		{
			name: "redis返回error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis错误"))
				res.EXPECT().Eval(gomock.Any(), luaSetCaptcha,
					[]string{keyFunc("test", "15212345678")},
					[]any{"123456"}).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: errors.New("redis错误"),
		},
		{
			name: "验证码不存在过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-2))
				res.EXPECT().Eval(gomock.Any(), luaSetCaptcha,
					[]string{keyFunc("test", "15212345678")},
					[]any{"123456"}).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: errors.New("system error"),
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-1))
				res.EXPECT().Eval(gomock.Any(), luaSetCaptcha,
					[]string{keyFunc("test", "15212345678")},
					[]any{"123456"}).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "15212345678",
			code:    "123456",
			wantErr: ErrSendTooFrequently,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewCaptchaCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			fmt.Println("err: ", err)
			fmt.Println("wanterr: ", tc.wantErr)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

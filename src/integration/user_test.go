package integration

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/solunara/isb/src/integration/startup"
	"github.com/solunara/isb/src/types/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed config.yaml
var defaultYAML string

func TestUserHandler_LoginSMSSend(t *testing.T) {
	server, redisCmd := startup.InitServer([]byte(defaultYAML))
	testCases := []struct {
		name string

		// 准备数据
		before func(t *testing.T)

		// 验证数据库数据
		after func(t *testing.T)

		reqBody string

		wantCode int
		wantBody app.ResponseType
	}{
		{
			name:   "发送验证码成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:user_login:15212345678"
				code, err := redisCmd.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) > 0)
				dur, err := redisCmd.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second+50)
				err = redisCmd.Del(ctx, key).Err()
				assert.NoError(t, err)
				key = "phone_code:user_login:15212345678:cnt"
				err = redisCmd.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			reqBody:  `{ "phone": "15212345678" }`,
			wantCode: 200,
			wantBody: app.ResponseOK(nil),
		},

		{
			name:     "未输入手机号码",
			before:   func(t *testing.T) {},
			after:    func(t *testing.T) {},
			reqBody:  `{ "phone": "" }`,
			wantCode: http.StatusOK,
			wantBody: *app.ErrEmptyRequest,
		},

		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:user_login:15212345679"
				err := redisCmd.Set(ctx, key, "123456", time.Minute*9+time.Second*50).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:user_login:15212345679"
				code, err := redisCmd.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
			},
			reqBody:  `{ "phone": "15212345679" }`,
			wantCode: http.StatusOK,
			wantBody: app.ResponseErr(400, "sent too often"),
		},

		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:user_login:15212345677"
				err := redisCmd.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:user_login:15212345677"
				code, err := redisCmd.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
				key = "phone_code:user_login:15212345677:cnt"
				err = redisCmd.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			reqBody:  `{ "phone": "15212345677" }`,
			wantCode: http.StatusOK,
			wantBody: *app.ErrInternalServer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost, "/user/login/sms/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			// 执行
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)

			var respData app.ResponseType
			err = json.NewDecoder(recorder.Body).Decode(&respData)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, respData)
			tc.after(t)
		})
	}
}

package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/service"
	svcmocks "github.com/solunara/isb/src/service/mocks"
	"github.com/solunara/isb/src/types/app"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUser_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		// mock
		mock func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService)

		// 构造请求，预期中输入
		reqBuilder func(t *testing.T) *http.Request

		// 预期中的输出
		wantCode int
		wantBody app.ResponseType
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), repository.User{
					Email:    "123@qq.com",
					Password: "hello#123456",
				}).Return(nil)
				codeSvc := svcmocks.NewMockCaptchaService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#123456",
"confirm_password": "hello#123456"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: app.ResponseOK("registration successful"),
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCaptchaService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
		"email": "123@",
		"password": "hello#world123",
		"confirm_password": "hello#world123"
		}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: *app.ErrBadRequestInvalidEmail,
		},
		{
			name: "两次密码输入不同",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCaptchaService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
		"email": "123@qq.com",
		"password": "hello#world123455",
		"confirm_password": "hello#world123"
		}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: *app.ErrBadRequestWrongPassword,
		},

		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCaptchaService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
		"email": "123@qq.com",
		"password": "hello",
		"confirm_password": "hello"
		}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: *app.ErrBadRequestInvalidPassword,
		},

		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), repository.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(app.ErrDuplicateEmail)
				codeSvc := svcmocks.NewMockCaptchaService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
		"email": "123@qq.com",
		"password": "hello#world123",
		"confirm_password": "hello#world123"
		}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: app.ResponseErr(400, app.ErrDuplicateEmail.Error()),
		},

		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CaptchaService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), repository.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("服务器内部错误"))
				codeSvc := svcmocks.NewMockCaptchaService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
		"email": "123@qq.com",
		"password": "hello#world123",
		"confirm_password": "hello#world123"
		}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusOK,
			wantBody: *app.ErrInternalServer,
		},
	}
	var respBody app.ResponseType
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构造 handler
			userSvc, codeSvc := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc)

			// 准备服务器，注册路由
			server := gin.Default()
			hdl.RegisterRoutes(server)

			// 准备Req和记录的 recorder
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()

			// 执行
			server.ServeHTTP(recorder, req)

			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)

			err := json.Unmarshal(recorder.Body.Bytes(), &respBody)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, respBody)
		})
	}
}

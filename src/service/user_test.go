package service

import (
	"context"
	"errors"
	"testing"

	"github.com/solunara/isb/src/repository"
	repomocks "github.com/solunara/isb/src/repository/mocks"
	"github.com/solunara/isb/src/types/app"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("123456#hello")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, []byte("123456#hello"))
	assert.NoError(t, err)
}

func TestUserService_LoginWithEmailPwd(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期输入
		ctx      context.Context
		email    string
		password string

		wantUser repository.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(repository.User{
						Email: "123@qq.com",
						// 你在这边拿到的密码，就应该是一个正确的密码
						// 加密后的正确的密码
						Password: "$2a$10$.l0JHmM7a2PdJ.A9gsmVyerEDlp1WhxsglC34S4UJH4TuHhWY7Tfq",
						Phone:    "15212345678",
					}, nil)
				return repo
			},
			email: "123@qq.com",
			// 用户输入的，没有加密的
			password: "123456#hello",

			wantUser: repository.User{
				Email:    "123@qq.com",
				Password: "$2a$10$.l0JHmM7a2PdJ.A9gsmVyerEDlp1WhxsglC34S4UJH4TuHhWY7Tfq",
				Phone:    "15212345678",
			},
		},

		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(repository.User{}, app.ErrRecordNotFound)
				return repo
			},
			email: "123@qq.com",
			// 用户输入的，没有加密的
			password: "123456#hello",
			wantErr:  app.ErrInvalidUserOrPassword,
		},

		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(repository.User{
						Email: "123@qq.com",
						// 你在这边拿到的密码，就应该是一个正确的密码
						// 加密后的正确的密码
						Password: "$2a$10$.l0JHmM7a2PdJ.A9gsmVyerEDlp1WhxsglC34S4UJH4TuHhWY7Tfq",
						Phone:    "15212345678",
					}, nil)
				return repo
			},
			email: "123@qq.com",
			// 用户输入的，没有加密的
			password: "123456#helloABCde",

			wantErr: app.ErrInvalidUserOrPassword,
		},

		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().
					FindByEmail(gomock.Any(), "123@qq.com").
					Return(repository.User{}, errors.New("db错误"))
				return repo
			},
			email: "123@qq.com",
			// 用户输入的，没有加密的
			password: "123456#hello",
			wantErr:  errors.New("db错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			user, err := svc.LoginWithEmailPwd(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

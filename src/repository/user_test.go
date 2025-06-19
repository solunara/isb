package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/repository/cache"
	cachemocks "github.com/solunara/isb/src/repository/cache/mocks"
	"github.com/solunara/isb/src/repository/dao"
	daomocks "github.com/solunara/isb/src/repository/dao/mocks"
	"github.com/solunara/isb/src/types/app"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO)

		ctx context.Context
		uid int64

		wantUser User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(model.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).
					Return(model.User{
						Id: uid,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						Profile:  "自我介绍",
						Phone: sql.NullString{
							String: "15212345678",
							Valid:  true,
						},
						Ctime: nowMs,
						Utime: nowMs,
					}, nil)
				c.EXPECT().Set(gomock.Any(), model.User{
					Id: 123,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: 100,
					Profile:  "自我介绍",
					Phone: sql.NullString{
						String: "15212345678",
						Valid:  true,
					},
					Ctime: nowMs,
					Utime: nowMs,
				}, time.Duration(time.Hour*240)).Return(nil)
				return c, d
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(100),
				Profile:  "自我介绍",
				Phone:    "15212345678",
			},
			wantErr: nil,
		},

		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(model.User{
						Id: uid,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						Profile:  "自我介绍",
						Phone: sql.NullString{
							String: "15212345678",
							Valid:  true,
						},
						Ctime: nowMs,
						Utime: nowMs,
					}, nil)
				return c, d
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(100),
				Profile:  "自我介绍",
				Phone:    "15212345678",
			},
			wantErr: nil,
		},

		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(model.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).
					Return(model.User{}, app.ErrRecordNotFound)
				return c, d
			},
			uid:      123,
			ctx:      context.Background(),
			wantUser: User{},
			wantErr:  app.ErrRecordNotFound,
		},

		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), uid).
					Return(model.User{}, cache.ErrKeyNotExist)
				d.EXPECT().FindById(gomock.Any(), uid).
					Return(model.User{
						Id: uid,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						Profile:  "自我介绍",
						Phone: sql.NullString{
							String: "15212345678",
							Valid:  true,
						},
						Ctime: 101,
						Utime: 102,
					}, nil)
				c.EXPECT().Set(gomock.Any(), model.User{
					Id: uid,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: 100,
					Profile:  "自我介绍",
					Phone: sql.NullString{
						String: "15212345678",
						Valid:  true,
					},
					Ctime: 101,
					Utime: 102,
				}, time.Duration(time.Hour*240)).Return(errors.New("redis错误"))
				return c, d
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: time.UnixMilli(100),
				Profile:  "自我介绍",
				Phone:    "15212345678",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			uc, ud := tc.mock(ctrl)
			rp := NewUserRepository(ud, uc)
			user, err := rp.FindById(tc.ctx, tc.uid)
			fmt.Println("err: ", err)
			fmt.Println("wanterr: ", tc.wantErr)
			fmt.Println("user: ", user)
			fmt.Println("wantuser: ", tc.wantUser)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

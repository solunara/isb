package dao

import (
	"context"
	"time"

	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/types/app"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type MsUserDAO interface {
	Insert(ctx context.Context, u model.MsUser) error
	FindById(ctx context.Context, uid int64) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.MsUser, error)
	FindByPhone(ctx context.Context, phone string) (model.User, error)
	UpdateUser(ctx context.Context, u model.User) (model.User, error)
}

type GORMMsUserDAO struct {
	db *gorm.DB
}

func NewMsUserDAO(db *gorm.DB) MsUserDAO {
	return &GORMMsUserDAO{
		db: db,
	}
}

func (dao *GORMMsUserDAO) Insert(ctx context.Context, u model.MsUser) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突，邮箱冲突
			return app.ErrDuplicateEmail
		}
	}
	return err
}

func (dao *GORMMsUserDAO) FindById(ctx context.Context, uid int64) (model.User, error) {
	var res model.User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&res).Error
	return res, err
}

func (dao *GORMMsUserDAO) FindByUsername(ctx context.Context, username string) (model.MsUser, error) {
	var u model.MsUser
	err := dao.db.WithContext(ctx).Where("username=?", username).First(&u).Error
	return u, err
}

func (dao *GORMMsUserDAO) FindByPhone(ctx context.Context, phone string) (model.User, error) {
	var res model.User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&res).Error
	return res, err
}

func (dao *GORMMsUserDAO) UpdateUser(ctx context.Context, u model.User) (model.User, error) {
	var res model.User
	err := dao.db.WithContext(ctx).Updates(&u).Error
	return res, err
}

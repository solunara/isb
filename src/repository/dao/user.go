package dao

import (
	"context"
	"time"

	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/types/app"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserDAO interface {
	Insert(ctx context.Context, u model.User) error
	FindById(ctx context.Context, uid int64) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByPhone(ctx context.Context, phone string) (model.User, error)
	FindByWechat(ctx context.Context, openID string) (model.User, error)
	UpdateUser(ctx context.Context, u model.User) (model.User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u model.User) error {
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

func (dao *GORMUserDAO) FindById(ctx context.Context, uid int64) (model.User, error) {
	var res model.User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&res).Error
	return res, err
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (model.User, error) {
	var res model.User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&res).Error
	return res, err
}

func (dao *GORMUserDAO) FindByWechat(ctx context.Context, openID string) (model.User, error) {
	var u model.User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openID).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateUser(ctx context.Context, u model.User) (model.User, error) {
	var res model.User
	err := dao.db.WithContext(ctx).Updates(&u).Error
	return res, err
}

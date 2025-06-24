package model

import (
	"database/sql"
)

const TableUser = "user"

func (User) TableName() string {
	return TableUser
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement" json:"id"`

	WechaOpenId sql.NullString `gorm:"unique" json:"wechat_open_id"`
	Phone       sql.NullString `gorm:"unique" json:"phone"`
	Email       sql.NullString `gorm:"unique" json:"email"`

	// encrypted password
	Password string `gorm:"type=varchar(256)" json:"password"`

	Nickname string `gorm:"type=varchar(128)" json:"nickname"`
	Profile  string `gorm:"type=varchar(4096)" json:"profile"`

	// unix time
	Birthday int64 `json:"birthday"`
	Ctime    int64 `json:"ctime"`
	Utime    int64 `json:"utime"`
}

type MsUser struct {
	Id    int64          `gorm:"primaryKey,autoIncrement" json:"id"`
	Phone sql.NullString `gorm:"unique" json:"phone"`
	Email sql.NullString `gorm:"unique" json:"email"`

	// encrypted password
	Password string `gorm:"type=varchar(256)" json:"password"`

	Username string `gorm:"type=varchar(32) unique" json:"username"`
	Profile  string `gorm:"type=varchar(4096)" json:"profile"`

	// unix time
	Birthday int64 `json:"birthday"`
	Ctime    int64 `json:"ctime"`
	Utime    int64 `json:"utime"`
}

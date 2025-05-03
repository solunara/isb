package xytmodel

import "database/sql"

const (
	TableXytUser = "xyt_user"
)

type XytUser struct {
	Id    int64          `gorm:"primaryKey,autoIncrement" json:"id"`
	Phone sql.NullString `gorm:"unique" json:"phone"`
	Email sql.NullString `gorm:"unique" json:"email"`

	// encrypted password
	Password string `gorm:"type=varchar(256)" json:"password"`

	Nickname string `gorm:"type=varchar(128)" json:"nickname"`
	Profile  string `gorm:"type=varchar(4096)" json:"profile"`

	// unix time
	Birthday int64 `json:"birthday"`
	Ctime    int64 `json:"ctime"`
	Utime    int64 `json:"utime"`
}

func (XytUser) TableName() string {
	return TableXytUser
}

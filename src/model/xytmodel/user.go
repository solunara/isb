package xytmodel

import (
	"database/sql"
	"time"
)

const (
	TableXytUser = "xyt_user"
)

type XytUser struct {
	Id     int64  `gorm:"primaryKey,autoIncrement" json:"id"`
	UserId string `gorm:"type=varchar(128);unique" json:"userId"`

	Email sql.NullString `gorm:"unique" json:"email"`
	Phone sql.NullString `gorm:"unique" json:"phone"`

	// encrypted password
	Password string `gorm:"type=varchar(256)" json:"password"`

	Name    string `gorm:"type=varchar(128)" json:"name"`
	Profile string `gorm:"type=varchar(4096)" json:"profile"`

	//
	IdNumber string `gorm:"type=varchar(24)" json:"idNumber"`

	Birthday string `gorm:"type=varchar(24)" json:"birthday"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (XytUser) TableName() string {
	return TableXytUser
}

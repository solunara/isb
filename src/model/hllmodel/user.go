package hllmodel

import (
	"database/sql"
	"time"
)

const (
	TableHllUser = "hll_user"
)

type HllUser struct {
	Id int64 `gorm:"primaryKey,autoIncrement" json:"id"`

	UserId   string `gorm:"type=varchar(128);unique;not null" json:"userId"`
	Username string `gorm:"type=varchar(32);unique" json:"username"`

	Email sql.NullString `gorm:"unique" json:"email"`
	Phone sql.NullString `gorm:"unique" json:"phone"`

	// encrypted password
	Password string `gorm:"type=varchar(256)" json:"password"`

	State string `gorm:"type=varchar(12)" json:"state"`

	Reserve int `gorm:"" json:"reserve"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (HllUser) TableName() string {
	return TableHllUser
}

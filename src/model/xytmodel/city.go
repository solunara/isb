package xytmodel

import "time"

const (
	TableProvince = "province"
	TableCity     = "city"
	TableDistrict = "district"
)

// 省表
type Province struct {
	Id        uint      `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name      string    `json:"name" gorm:"column:name;type:varchar(32);not null;comment:省|自治区|直辖市名称"`
	Abbr_zh   string    `json:"abbr_zh" gorm:"column:abbr_zh;type:varchar(4);not null;comment:中文简称"`
	Abbr_en   string    `json:"abbr_en" gorm:"column:abbr_en;type:varchar(2);not null;comment:英文简称"`
	Code      string    `json:"code" gorm:"column:code;size:6;comment:行政区划编码"`
	Category  uint8     `json:"category" gorm:"column:category;comment:类别: 1:省 2:自治区 3:直辖市"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 市表
type City struct {
	Id           uint   `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name         string `json:"name" gorm:"column:name;type:varchar(64);not null;comment:市名称"`
	Code         string `json:"code" gorm:"column:code;size:6;comment:市行政区划编码"`
	ProvinceName string `json:"province_name" gorm:"column:province_name;type:varchar(32);comment:省名称"`
	ProvinceCode string `json:"province_code" gorm:"column:province_code;size:6;comment:省份行政区划编码"`
	Category     uint8  `json:"category" gorm:"column:category;comment:类别: 1:普通省份城市 2:自治区 3:直辖市"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// 区县表
type District struct {
	Id           uint   `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name         string `json:"name" gorm:"column:name;type:varchar(64);not null;comment:区县名称"`
	Code         string `json:"code" gorm:"column:code;size:12;comment:区县行政区划编码"`
	CityName     string `json:"city_name" gorm:"column:city_name;type:varchar(64);comment:市名称"`
	CityCode     string `json:"city_code" gorm:"column:city_code;size:6;comment:市行政区划编码"`
	ProvinceName string `json:"province_name" gorm:"column:province_name;type:varchar(32);comment:省名称"`
	ProvinceCode string `json:"province_code" gorm:"column:province_code;size:6;comment:省份行政区划编码"`
	Category     uint8  `json:"category" gorm:"column:category;comment:类别: 1:普通省份区县 2:自治区 3:直辖市"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

func (Province) TableName() string {
	return TableProvince
}

func (City) TableName() string {
	return TableCity
}

func (District) TableName() string {
	return TableDistrict
}

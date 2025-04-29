package xytmodel

const (
	TableHospital      = "hospital"
	TableHospitalGrade = "hospital_grade"
	TableCity          = "city"
	TableDistrict      = "district"
)

type Hospital struct {
	HospitalID          string  `json:"hospital_id" gorm:"column:hospital_id;primaryKey;size:64;not null;comment:医院唯一编码"`
	HospitalName        string  `json:"hospital_name" gorm:"column:hospital_name;size:128;not null;comment:医院全称"`
	HospitalShortName   string  `json:"hospital_short_name" gorm:"column:hospital_short_name;size:64;comment:医院简称"`
	HospitalTypeCode    string  `json:"hospital_type_code" gorm:"column:hospital_type_code;size:10;comment:医院类型编码"`
	HospitalTypeName    string  `json:"hospital_type_name" gorm:"column:hospital_type_name;size:64;comment:医院类型名称"`
	HospitalGradeCode   string  `json:"hospital_grade_code" gorm:"column:hospital_grade_code;size:10;comment:医院等级编码"`
	HospitalGradeName   string  `json:"hospital_grade_name" gorm:"column:hospital_grade_name;size:64;comment:医院等级名称"`
	EconomicTypeCode    string  `json:"economic_type_code" gorm:"column:economic_type_code;size:10;comment:经济类型编码"`
	EconomicTypeName    string  `json:"economic_type_name" gorm:"column:economic_type_name;size:64;comment:经济类型名称"`
	ProvinceCode        string  `json:"province_code" gorm:"column:province_code;size:6;comment:省份编码"`
	ProvinceName        string  `json:"province_name" gorm:"column:province_name;size:64;comment:省份名称"`
	CityCode            string  `json:"city_code" gorm:"column:city_code;size:6;comment:城市编码"`
	CityName            string  `json:"city_name" gorm:"column:city_name;size:64;comment:城市名称"`
	DistrictCode        string  `json:"district_code" gorm:"column:district_code;size:6;comment:区县编码"`
	DistrictName        string  `json:"district_name" gorm:"column:district_name;size:64;comment:区县名称"`
	Address             string  `json:"address" gorm:"column:address;size:256;comment:详细地址"`
	Telephone           string  `json:"telephone" gorm:"column:telephone;size:50;comment:联系电话"`
	WebsiteURL          string  `json:"website_url" gorm:"column:website_url;size:256;comment:医院官网"`
	LegalRepresentative string  `json:"legal_representative" gorm:"column:legal_representative;size:64;comment:法定代表人"`
	OrgCode             string  `json:"org_code" gorm:"column:org_code;size:64;comment:组织机构代码"`
	LicenseNumber       string  `json:"license_number" gorm:"column:license_number;size:64;comment:医疗机构执业许可证号"`
	LogoUrl             string  `json:"logo_url" gorm:"column:logo_url;size:128;comment:logo图片地址"`
	Longitude           float64 `json:"longitude" gorm:"column:longitude;type:decimal(10,6);comment:经度"`
	Latitude            float64 `json:"latitude" gorm:"column:latitude;type:decimal(10,6);comment:纬度"`
	LicenseExpiry       int64   `json:"license_expiry" gorm:"column:license_expiry;type:date;comment:许可证有效期"`
	EstablishedAt       int64   `json:"established_at" gorm:"type:date;comment:成立时间"`
	CreatedAt           int64   `json:"created_at" gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt           int64   `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	IsMedicalInsurance  bool    `json:"is_medical_insurance" gorm:"column:is_medical_insurance;comment:是否支持医保"`
	IsActive            bool    `json:"is_active" gorm:"column:is_active;default:true;comment:是否启用"`
}

type HospitalGrade struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:自增主键"`
	GradeCode string `json:"grade_code" gorm:"column:grade_code;size:10;comment:医院等级编码"`
	GradeName string `json:"grade_name" gorm:"column:grade_name;size:64;comment:医院等级名称"`
}

type City struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	CityName     string `json:"city_name" gorm:"column:city_name;type:varchar(100);not null;comment:城市名称"`
	ProvinceName string `json:"province_name" gorm:"column:province_name;type:varchar(100);comment:省份名称"`
	CityCode     string `json:"city_code" gorm:"column:city_code;size:6;comment:城市编码"`
	ProvinceCode string `json:"province_code" gorm:"column:province_code;size:6;comment:省份编码"`
	CreatedAt    int64
	UpdatedAt    int64
}

type District struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	DistrictName string `json:"district_name" gorm:"column:district_name;type:varchar(100);not null;comment:区名称"`
	DistrictCode string `json:"district_code" gorm:"column:district_code;size:6;comment:区编码"`
	CityCode     string `json:"city_code" gorm:"column:city_code;size:6;comment:城市编码"`
	ProvinceCode string `json:"province_code" gorm:"column:province_code;size:6;comment:省份编码"`
	CreatedAt    int64
	UpdatedAt    int64
}

func (Hospital) TableName() string {
	return TableHospital
}

func (HospitalGrade) TableName() string {
	return TableHospitalGrade
}

func (City) TableName() string {
	return TableCity
}

func (District) TableName() string {
	return TableDistrict
}

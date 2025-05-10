package xytmodel

import "time"

const (
	TableHospital      = "hospital"
	TableHospitalGrade = "hospital_grade"
	TableDepartment    = "department"
	TableSchedule      = "schedule"
	TableDoctor        = "doctor"
	TablePatient       = "patient"
	TableOrder         = "register_order"
)

// 医院表
type Hospital struct {
	UID                 string  `json:"uid" gorm:"column:uid;primaryKey;size:64;not null;comment:医院唯一编码/卫健委登记号"`
	FullName            string  `json:"full_name" gorm:"column:full_name;size:128;not null;comment:医院全称"`
	ShortName           string  `json:"short_name" gorm:"column:short_name;size:64;comment:医院简称"`
	TypeCode            string  `json:"type_code" gorm:"column:type_code;size:10;comment:医院类型编码"`
	TypeName            string  `json:"type_name" gorm:"column:type_name;size:64;comment:医院类型名称"`
	GradeCode           string  `json:"grade_code" gorm:"column:grade_code;size:10;comment:医院等级编码"`
	GradeName           string  `json:"grade_name" gorm:"column:grade_name;size:64;comment:医院等级名称"`
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
	LicenseExpiry       string  `json:"license_expiry" gorm:"column:license_expiry;size:12;comment:许可证有效期"`
	EstablishedAt       string  `json:"established_at" gorm:"size:12;comment:成立时间"`
	CreatedAt           int64   `json:"created_at" gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt           int64   `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	IsMedicalInsurance  bool    `json:"is_medical_insurance" gorm:"column:is_medical_insurance;comment:是否支持医保"`
	IsActive            bool    `json:"is_active" gorm:"column:is_active;default:true;comment:是否启用"`
}

// 医院等级表
type HospitalGrade struct {
	Id        uint   `json:"id" gorm:"primaryKey;autoIncrement;column:id;comment:自增主键"`
	GradeCode string `json:"grade_code" gorm:"column:grade_code;size:10;comment:医院等级编码"`
	GradeName string `json:"grade_name" gorm:"column:grade_name;size:64;comment:医院等级名称"`
}

// 科室表
type Department struct {
	UID         string `json:"uid" gorm:"column:uid;primaryKey;size:64;not null;comment:科室唯一编码"`
	HospitalID  string `gorm:"not null;index;size:64;" json:"hospital_id"` // 所属医院
	Name        string `gorm:"type:varchar(100);not null" json:"name"`     // 科室名称
	Description string `gorm:"type:text" json:"description"`               // 科室简介
	ParentID    string `gorm:"size:64;index" json:"parent_id"`             // 上级科室ID，空表示顶级科室
	Level       int    `gorm:"type:int;default:1" json:"level"`            // 科室层级
	Status      bool   `gorm:"type:tinyint(1);default:1" json:"status"`    // 启用状态 true=启用，false=停用
	SortOrder   int    `gorm:"type:int;default:0" json:"sort_order"`       // 排序字段
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// 医生表
type Doctor struct {
	Id        string `gorm:"column:id;size:24;primaryKey" json:"id"`
	Name      string `gorm:"column:name;size:50;not null" json:"name"`
	Rank      string `gorm:"column:rank;size:50" json:"rank"`
	DeptId    string `gorm:"column:dept_id;size:64;not null" json:"deptId"`
	HosId     string `gorm:"column:hos_id;size:24;not null" json:"hosId"`
	Profile   string `gorm:"column:profile;size:300" json:"profile"`
	Phone     string `gorm:"column:phone;size:15" json:"phone"`
	Email     string `gorm:"column:email;size:100" json:"email"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// 挂号类型表
type RegistrationType struct {
	Id          uint    `gorm:"primaryKey;autoIncrement;" json:"id"`
	Name        string  `gorm:"size:50;not null" json:"name"`
	Fee         float64 `gorm:"type:decimal(10,2);default:0.00" json:"fee"`
	Description string  `gorm:"type:text" json:"description"`
}

// 挂号表
type Registration struct {
	Id        uint      `gorm:"primaryKey" json:"id"`
	PatientId uint      `gorm:"not null" json:"patientId"`
	DocId     uint      `gorm:"not null;size:24;" json:"docId"`
	HosID     string    `gorm:"column:hos_id;size:24;" json:"hosId"`
	DeptID    string    `gorm:"column:dept_id;size:64;" json:"deptId"`
	RegType   uint      `gorm:"not null" json:"regType"`
	RegTime   string    `gorm:"not null;size:32;" json:"regTime"`
	VisitTime string    `gorm:"not null;size:32;" json:"visitTime,omitempty"`
	Status    string    `gorm:"type:enum('已挂号','已就诊','取消');default:'已挂号'" json:"status"`
	Fee       float64   `gorm:"type:decimal(10,2);default:0.00" json:"fee"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 排班表
type Schedule struct {
	Id          int    `gorm:"column:id;primaryKey" json:"id"`
	ScheId      string `gorm:"column:sche_id;not null;size:24;" json:"scheId"`
	DocId       string `gorm:"column:doc_id;not null;size:24;" json:"docId"`
	HosID       string `gorm:"column:hos_id;size:24;" json:"hosId"`
	DeptID      string `gorm:"column:dept_id;size:64;" json:"deptId"`
	WorkDate    string `gorm:"column:work_date;size:12;" json:"workDate"`
	TimeSlot    string `gorm:"column:time_slot;type:enum('上午','下午','晚上');not null" json:"timeSlot"`
	Amount      int    `gorm:"column:amount;not null" json:"amount"`
	WorkWeek    int    `gorm:"column:work_week;not null" json:"workWeek"`
	MaxPatients int    `gorm:"column:max_patients;default:20" json:"maxPatients"`
	Registered  int    `gorm:"column:registered;default:0" json:"registered"`
}

// 就诊人表
type Patient struct {
	Id                       string    `gorm:"column:id;primaryKey;size:64;common:就诊人唯一id" json:"id"`
	Name                     string    `gorm:"column:name;not null;size:64;common:就诊人姓名" json:"name"`
	UserId                   string    `gorm:"column:user_id;not null;size:64;common:就诊人所属用户id" json:"userId"`
	ProvinceCode             string    `gorm:"column:province_code;not null;size:6;common:所在省份编码" json:"provinceCode"`
	CityCode                 string    `gorm:"column:city_code;not null;size:6;common:所在市编码" json:"cityCode"`
	DistrictCode             string    `gorm:"column:district_code;not null;size:6;common:所在区县编码" json:"districtCode"`
	CertificatesNo           string    `gorm:"column:certificates_no;not null;size:24;common:实名号" json:"certificatesNo"`
	Address                  string    `gorm:"column:address;not null;size:128;common:详细地址" json:"address"`
	ContactsName             string    `gorm:"column:contacts_name;size:64;common:联系人姓名" json:"contactsName"`
	ContactsCertificatesNo   string    `gorm:"column:contacts_certificates_no;size:24;common:联系人实名号" json:"contactsCertificatesNo"`
	ContactsPhone            string    `gorm:"column:contacts_phone;size:16;common:联系人手机号" json:"contactsPhone"`
	Birthday                 string    `gorm:"column:birthday;size:20;" json:"birthday"`
	Phone                    string    `gorm:"column:phone;not null;size:16;" json:"phone"`
	CertificatesType         uint8     `gorm:"column:certificates_type;not null;common:实名认证类型" json:"certificatesType"`            // 0: 身份证 1:户口本
	ContactsCertificatesType uint8     `gorm:"column:contacts_certificates_type;common:联系人实名认证类型" json:"contactsCertificatesType"` // 0: 身份证 1:户口本
	Sex                      uint8     `gorm:"column:sex;not null;common:性别" json:"sex"`                                           // 0:女性 1: 男性
	IsMarry                  uint8     `gorm:"column:is_marry;common:是否已婚" json:"isMarry"`                                         // 0: 未婚 1:已婚
	IsInsure                 uint8     `gorm:"column:is_insure;common:是否是医保用户" json:"isInsure"`                                    // 0: 医保 1:自费
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// 挂号订单表
type RegisterOrder struct {
	Id           int       `gorm:"column:id;primaryKey" json:"id"`
	UserId       string    `gorm:"column:user_id;not null;size:64;" json:"userId"`
	OrderId      string    `gorm:"column:order_id;not null;size:64;unique;" json:"orderId"`
	PatientId    string    `gorm:"column:patient_id;not null;size:64;" json:"patientId"`
	HosID        string    `gorm:"column:hos_id;size:24;" json:"hosId"`
	DeptID       string    `gorm:"column:dept_id;size:64;" json:"deptId"`
	DocId        string    `gorm:"column:doc_id;not null;size:24;" json:"docId"`
	HosName      string    `gorm:"column:hos_name;size:128;not null;" json:"hosName"`
	DeptName     string    `gorm:"column:dept_name;size:32;not null;" json:"deptName"`
	DocName      string    `gorm:"column:doc_name;size:24;not null;" json:"docName"`
	PatientName  string    `gorm:"column:patient_name;size:24;not null;" json:"patientName"`
	VisitTime    string    `gorm:"column:visit_time;not null;size:24;" json:"visitTime"`
	Amount       int       `gorm:"column:amount;not null" json:"amount"`
	State        int8      `gorm:"column:state;" json:"state"` // -1: 已取消  0: 待支付  1:已支付  2:已完成
	RegisterTime string    `gorm:"column:register_time;not null;size:24;" json:"registerTime"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// 挂号订单表
type RegisterOrder struct {
	Id           int       `gorm:"column:id;primaryKey" json:"id"`
	UserId       string    `gorm:"column:user_id;not null;size:64;unique;" json:"userId"`
	OrderId      string    `gorm:"column:order_id;not null;size:64;unique;" json:"orderId"`
	PatientId    string    `gorm:"column:patient_id;not null;size:64;" json:"patientId"`
	HosID        string    `gorm:"column:hos_id;size:24;" json:"hosId"`
	DeptID       string    `gorm:"column:dept_id;size:64;" json:"deptId"`
	DocId        string    `gorm:"column:doc_id;not null;size:24;" json:"docId"`
	HosName      string    `gorm:"column:hos_name;size:128;not null;" json:"hosName"`
	DeptName     string    `gorm:"column:dept_name;size:32;not null;" json:"deptName"`
	DocName      string    `gorm:"column:doc_name;size:24;not null;" json:"docName"`
	PatientName  string    `gorm:"column:patient_name;size:24;not null;" json:"patientName"`
	VisitTime    string    `gorm:"column:visit_time;not null;size:24;" json:"visitTime"`
	Amount       int       `gorm:"column:amount;not null" json:"amount"`
	State        int8      `gorm:"column:state;" json:"state"` // -1: 已取消  0: 待支付  1:已支付  2:已完成
	RegisterTime string    `gorm:"column:register_time;not null;size:24;" json:"registerTime"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Hospital) TableName() string {
	return TableHospital
}

func (HospitalGrade) TableName() string {
	return TableHospitalGrade
}

func (Department) TableName() string {
	return TableDepartment
}

func (Schedule) TableName() string {
	return TableSchedule
}

func (Doctor) TableName() string {
	return TableDoctor
}

func (Patient) TableName() string {
	return TablePatient
}

func (RegisterOrder) TableName() string {
	return TableOrder
}

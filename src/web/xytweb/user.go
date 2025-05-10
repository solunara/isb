package xytweb

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/config"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
	"github.com/solunara/isb/src/utils"
	"gorm.io/gorm"
)

type XytUserHandler struct {
	cache redis.Cmdable
	db    *gorm.DB
}

func NewXytUserlHandler(cache redis.Cmdable, db *gorm.DB) *XytUserHandler {
	return &XytUserHandler{
		cache: cache,
		db:    db,
	}
}

func (xh *XytUserHandler) RegisterRoutes(group *gin.RouterGroup) {
	// ---------------- vbook api ---------------------
	ug := group.Group("/user")
	ug.GET("/phone/code", xh.phoneCode)
	ug.POST("/login/phone", xh.loginByPhone)
	ug.GET("/login/wechat/param", xh.wechatParam)
	ug.GET("/info", xh.getUser)
	ug.POST("/certification", xh.certification)
	ug.GET("/patient/list", xh.getPatients)
	ug.GET("/order/states", xh.getOrderStates)
	ug.GET("/order/list", xh.getOrderList)
	ug.POST("/patient", xh.addPatient)
	ug.PUT("/patient", xh.updatePatient)
}

type AddOrUpdateUser struct {
	Id                       string   `json:"id"`
	Name                     string   `json:"name"`
	CertificatesType         string   `json:"certificatesType"`
	CertificatesNo           string   `json:"certificatesNo"`
	Sex                      int      `json:"sex"`
	Birthdate                string   `json:"birthdate"`
	Phone                    string   `json:"phone"`
	IsMarry                  int      `json:"isMarry"`
	IsInsure                 int      `json:"isInsure"`
	AddressSelected          []string `json:"addressSelected"`
	Address                  string   `json:"address"`
	ContactsName             string   `json:"contactsName"`
	ContactsCertificatesType string   `json:"contactsCertificatesType"`
	ContactsCertificatesNo   string   `json:"contactsCertificatesNo"`
	ContactsPhone            string   `json:"contactsPhone"`
}

type ViewUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Profile  string `json:"profile"`
	IdNumber string `json:"idNumber"`
	Birthday string `json:"birthday"`
}

type OrderStatus struct {
	Id    int    `json:"id"`
	State string `json:"state"`
}

func (xh *XytUserHandler) addPatient(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	var req AddOrUpdateUser
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	err := CreatePatient(xh.db, userid.(string), req)
	if err != nil {
		if errors.Is(err, app.ErrUserNotFound) {
			ctx.JSON(http.StatusOK, app.ErrNotFound)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(nil))
}

func (xh *XytUserHandler) updatePatient(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	var req AddOrUpdateUser
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	err := UpdatePatient(xh.db, userid.(string), req)
	if err != nil {
		if errors.Is(err, app.ErrUserNotFound) {
			ctx.JSON(http.StatusOK, app.ErrNotFound)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(nil))
}

func (xh *XytUserHandler) getOrderList(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	queries := ctx.Request.URL.Query()
	dbQuery := xh.db.Model(&xytmodel.RegisterOrder{})
	dbQuery.Where("user_id = ?", userid.(string))
	for k, v := range queries {
		switch k {
		case "patientId":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("patient_id = ?", v[0])
			}
		case "state":
			if len(v) > 0 && v[0] != "" {
				st, err := strconv.Atoi(v[0])
				if err == nil {
					if st == -1 || st == 0 || st == 1 || st == 2 {
						dbQuery.Where("state = ?", v[0])
					}
				}
			}
		case "pageNo", "pageSize", "timeStamp":
			continue
		}
	}

	pageNo, _ := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "1"))
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > 30 {
		pageSize = 30
	}
	offset := (pageNo - 1) * pageSize
	var total int64
	err := dbQuery.Count(&total).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	if offset > int(total) {
		ctx.JSON(200, app.ErrOutOfRange)
		return
	}

	if total < 1 {
		ctx.JSON(http.StatusOK, app.ResponsePageData(0, []xytmodel.RegisterOrder{}))
		return
	}

	var orderlist []xytmodel.RegisterOrder
	err = dbQuery.Limit(pageSize).Offset(offset).Find(&orderlist).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponsePageData(total, orderlist))
}

func (xh *XytUserHandler) getOrderStates(ctx *gin.Context) {
	_, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK([]OrderStatus{
		{-1, "已取消"},
		{0, "待支付"},
		{1, "已支付"},
		{2, "已完成"},
	}))
}

func (xh *XytUserHandler) getPatients(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	var patients []xytmodel.Patient
	err := xh.db.Table(xytmodel.TablePatient).Where("user_id = ?", userid.(string)).Find(&patients).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(patients))
}

type CertificationReq struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	CodeType string `json:"codeType"`
	Image    string `json:"image"`
}

func (xh *XytUserHandler) certification(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	var req CertificationReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	var xytuser xytmodel.XytUser
	err := xh.db.Table(xytmodel.TableXytUser).Where("user_id = ?", userid.(string)).Take(&xytuser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, app.ErrNotFound)
			return
		}
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	if xytuser.IdNumber != "" && xytuser.IdTyper != "" {
		ctx.JSON(http.StatusOK, app.ResponseOK(true))
		return
	}

	err = xh.db.Table(xytmodel.TableXytUser).Where("user_id = ?", userid.(string)).Updates(&xytmodel.XytUser{
		Name:     req.Name,
		IdNumber: req.Code,
		IdTyper:  req.CodeType,
		Image:    []byte(req.Image),
	}).Error
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(true))
}

func (xh *XytUserHandler) getUser(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	var xytuser xytmodel.XytUser
	err := xh.db.Table(xytmodel.TableXytUser).Where("user_id = ?", userid.(string)).Take(&xytuser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, app.ErrNotFound)
			return
		}
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(ViewUser{
		Name:     xytuser.Name,
		Email:    xytuser.Email.String,
		Phone:    xytuser.Phone.String,
		Profile:  xytuser.Profile,
		IdNumber: xytuser.IdNumber,
		Birthday: xytuser.Birthday,
	}))
}

func (xh *XytUserHandler) phoneCode(ctx *gin.Context) {
	var err error

	phone := ctx.Query("phone")
	if phone == "" {
		ctx.JSON(200, app.ResponseErr(400, "请输入手机号"))
		return
	}
	// 确保是六位数，通过格式化实现
	codeStr := fmt.Sprintf("%06d", rand.Intn(1000000))
	err = xh.cache.Set(ctx, phone, codeStr, time.Minute).Err()
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(codeStr))
}

func (xh *XytUserHandler) loginByPhone(ctx *gin.Context) {
	type LoginByPhoneReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req LoginByPhoneReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	code, err := xh.cache.Get(ctx, req.Phone).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			ctx.JSON(http.StatusOK, app.ErrBadPhoneOrCode)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	if code != req.Code {
		ctx.JSON(http.StatusOK, app.ErrBadPhoneOrCode)
		return
	}

	xytuser, err := FindOrCreateByPhone(xh.db, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	tokenVal, err := jwtoken.NewJWToken().CreateJWToken(jwtoken.CustomClaims{
		Name:   xytuser.Name,
		UserId: xytuser.UserId,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(map[string]string{"name": xytuser.Name, "userId": xytuser.UserId, "token": tokenVal}))
}

func (xh *XytUserHandler) wechatParam(ctx *gin.Context) {
	// TOTO: 获取微信扫码登录参数
	type RespData struct {
		RedirectUri string `json:"redirectUri"`
		Appid       string `json:"appid"`
		Scope       string `json:"scope"`
		State       string `json:"state"`
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(RespData{}))
}

func FindOrCreateByPhone(db *gorm.DB, phone string) (xytmodel.XytUser, error) {
	var xytuser xytmodel.XytUser
	err := db.Table(xytmodel.TableXytUser).Where("phone = ?", phone).Take(&xytuser).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return xytmodel.XytUser{}, err
		}
		xytuser.UserId = uuid.New().String()
		xytuser.Name = fmt.Sprintf("用户_%s****%s", phone[:3], phone[len(phone)-4:])
		xytuser.Phone = sql.NullString{
			String: phone,
			Valid:  true,
		}
		err = db.Create(&xytuser).Error
		if err != nil {
			return xytmodel.XytUser{}, err
		}
		return xytuser, nil
	}
	return xytuser, nil
}

func CreatePatient(db *gorm.DB, userId string, data AddOrUpdateUser) error {
	var sex bool
	if data.Sex == 1 {
		sex = true
	}
	var xytpatient = xytmodel.Patient{
		UserId:    userId,
		PatientId: uuid.NewString(),
		Name:      data.Name,
		Birthday:  data.Birthdate,
		Phone:     data.Phone,
		Sex:       sex,
	}
	var user xytmodel.XytUser
	err := db.Table(xytmodel.TableXytUser).Where("user_id = ?", userId).Take(&user).Error
	if err != nil {
		return app.ErrUserNotFound
	}
	err = db.Table(xytmodel.TablePatient).Create(&xytpatient).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
		xytpatient.UserId = uuid.NewString()
		return db.Table(xytmodel.TablePatient).Create(&xytpatient).Error
	}
	return nil
}

func UpdatePatient(db *gorm.DB, userId string, data AddOrUpdateUser) error {
	var sex bool
	if data.Sex == 1 {
		sex = true
	}

	var user xytmodel.XytUser
	err := db.Table(xytmodel.TableXytUser).Where("user_id = ?", userId).Take(&user).Error
	if err != nil {
		return app.ErrUserNotFound
	}
	var patient xytmodel.Patient
	err = db.Table(xytmodel.TablePatient).Where("patient_id = ?", data.Id).Take(&patient).Error
	if err != nil {
		return app.ErrUserNotFound
	}
	return db.Table(xytmodel.TablePatient).Where("patient_id = ?", data.Id).Updates(&xytmodel.Patient{
		Name:     data.Name,
		Birthday: data.Birthdate,
		Phone:    data.Phone,
		Sex:      sex,
	}).Error
}

func CreateOrder(db *gorm.DB, data AddOrderReq) (string, error) {
	var patient xytmodel.Patient
	err := db.Table(xytmodel.TablePatient).Where("patient_id = ?", data.PatientId).Take(&patient).Error
	if err != nil {
		return "", err
	}

	var schedule xytmodel.Schedule
	err = db.Table(xytmodel.TableSchedule).Where("sche_id = ?", data.ScheId).Take(&schedule).Error
	if err != nil {
		return "", err
	}

	var hospital xytmodel.Hospital
	err = db.Table(xytmodel.TableHospital).Where("uid = ?", schedule.HosID).Take(&hospital).Error
	if err != nil {
		return "", err
	}

	var dept xytmodel.Department
	err = db.Table(xytmodel.TableDepartment).Where("uid = ?", schedule.DeptID).Take(&dept).Error
	if err != nil {
		return "", err
	}

	var doctor xytmodel.Doctor
	err = db.Table(xytmodel.TableDoctor).Where("id = ?", schedule.DocId).Take(&doctor).Error
	if err != nil {
		return "", err
	}

	var xytorder = xytmodel.RegisterOrder{
		UserId:       patient.UserId,
		OrderId:      utils.GenerateUinqueID(),
		PatientId:    data.PatientId,
		HosID:        schedule.HosID,
		DeptID:       schedule.DeptID,
		DocId:        schedule.DocId,
		HosName:      hospital.FullName,
		DeptName:     dept.Name,
		DocName:      doctor.Name,
		PatientName:  patient.Name,
		VisitTime:    schedule.WorkDate + " " + schedule.TimeSlot,
		Amount:       schedule.Amount,
		State:        0,
		RegisterTime: time.Now().Format(time.DateTime),
	}

	err = db.Table(xytmodel.TableOrder).Create(&xytorder).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", err
		}
		xytorder.OrderId = utils.GenerateUinqueID()
		return xytorder.OrderId, db.Table(xytmodel.TableOrder).Create(&xytorder).Error
	}
	return xytorder.OrderId, nil
}

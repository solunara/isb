package xytweb

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/config"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/utils"
	"gorm.io/gorm"
)

type XytHospitalHandler struct {
	db *gorm.DB
}

const MaxPatientsPerDay = 10
const MaxSchedulerDays = 7

func NewXytHospitalHandler(db *gorm.DB) *XytHospitalHandler {
	return &XytHospitalHandler{
		db: db,
	}
}

func (xh *XytHospitalHandler) RegisterRoutes(group *gin.RouterGroup) {
	// ---------------- vbook api ---------------------
	ug := group.Group("/hos")
	ug.GET("/list", xh.hosList)
	ug.GET("/grade", xh.hosGrade)
	ug.GET("/region", xh.hosRegion)
	ug.GET("/detail", xh.hosDetail)
	ug.GET("/department", xh.hosDepartment)
	ug.GET("/scheduler", xh.docSchedules)
	ug.GET("/register/doctor", xh.getDoctor)
	ug.POST("/add/order", xh.addOrder)
	ug.GET("/order", xh.getOrder)
	ug.POST("/cancel/order", xh.cancelOrder)
	ug.GET("/order/list", xh.listOrder)
}

func (xh *XytHospitalHandler) hosList(ctx *gin.Context) {
	var err error
	var hoslist []xytmodel.Hospital
	queries := ctx.Request.URL.Query()
	dbQuery := xh.db.Model(&xytmodel.Hospital{})
	for k, v := range queries {
		switch k {
		case "hosId":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("uid = ?", v[0])
			}
		case "gradeCode":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("grade_code = ?", v[0])
			}
		case "cityCode":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("city_code = ?", v[0])
			}
		case "cityName":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("city_name = ?", v[0])
			}
		case "hosName":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("full_name like ?", "%"+v[0]+"%")
			}
		case "districtCode":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("district_code = ?", v[0])
			}
		case "pageNo", "pageSize", "timeStamp":
			continue
		}
	}

	v, ok := queries["hosName"]
	if ok {
		if len(v) > 0 && v[0] != "" {
			err = dbQuery.Find(&hoslist).Error
			if err != nil {
				ctx.JSON(200, app.ErrInternalServer)
				return
			}
			ctx.JSON(http.StatusOK, app.ResponseOK(hoslist))
			return
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
	err = dbQuery.Count(&total).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	if offset > int(total) {
		ctx.JSON(200, app.ResponsePageData(total, nil))
		return
	}

	if total < 1 {
		ctx.JSON(http.StatusOK, app.ResponsePageData(0, nil))
		return
	}

	err = dbQuery.Limit(pageSize).Offset(offset).Find(&hoslist).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponsePageData(total, hoslist))
}

func (xh *XytHospitalHandler) hosGrade(ctx *gin.Context) {
	dbQuery := xh.db.Table(xytmodel.TableHospitalGrade)
	var hosgrade []xytmodel.HospitalGrade
	err := dbQuery.Find(&hosgrade).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(hosgrade))
}

func (xh *XytHospitalHandler) hosRegion(ctx *gin.Context) {
	var err error
	var city_code string
	cityName := ctx.Query("cityName")
	if cityName == "" {
		cityCode := ctx.Param("cityCode")
		if cityCode == "" {
			ctx.JSON(200, app.ResponseErr(400, "请指定城市名字或编码"))
			return
		}
		city_code = cityCode
	} else {
		var city xytmodel.City
		unescapeCityName, err := url.QueryUnescape(cityName)
		if err == nil {
			cityName = unescapeCityName
		}
		switch cityName {
		case "北京市", "上海市", "天津市", "重庆市", "北京", "上海", "天津", "重庆":
			err = xh.db.Table(xytmodel.TableCity).Where("province_name = ?", cityName).Take(&city).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.JSON(200, app.ResponseErr(404, "请指定一个存在的城市名字"))
					return
				}
				ctx.JSON(200, app.ErrInternalServer)
				return
			}
			city_code = city.Code
		default:
			err = xh.db.Table(xytmodel.TableCity).Where("name = ?", cityName).Take(&city).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.JSON(200, app.ResponseErr(404, "请指定一个存在的城市名字"))
					return
				}
				ctx.JSON(200, app.ErrInternalServer)
				return
			}
			city_code = city.Code
		}
	}

	var district []xytmodel.District
	err = xh.db.Table(xytmodel.TableDistrict).Where("city_code = ?", city_code).Find(&district).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(district))
}

func (xh *XytHospitalHandler) hosDetail(ctx *gin.Context) {
	var err error
	uid := ctx.Query("hosId")
	if uid == "" {
		ctx.JSON(200, app.ResponseErr(400, "请指定医院hosId"))
		return
	}

	var hos xytmodel.Hospital
	err = xh.db.Table(xytmodel.TableHospital).Where("uid = ?", uid).Take(&hos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, app.ResponseErr(404, "找不到该医院"))
			return
		}
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(hos))
}

func (xh *XytHospitalHandler) hosDepartment(ctx *gin.Context) {
	var err error
	// uid := ctx.Query("uid")
	// if uid == "" {
	// 	ctx.JSON(200, app.ResponseErr(400, "请指定医院uid"))
	// 	return
	// }

	var department []xytmodel.Department
	err = xh.db.Table(xytmodel.TableDepartment).Order("sort_order").Find(&department).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(BuildDepartmentTree(department)))
}

type DocScheduler struct {
	DocId       string `json:"docId"`
	ScheId      string `json:"scheId"`
	TimeSlot    string `json:"timeSlot"`
	DoctorName  string `json:"doctorName"`
	Rank        string `json:"rank"`
	Profile     string `json:"profile"`
	WorkDay     string `json:"workDay"`
	Amount      int    `json:"amount"`
	MaxPatients int    `json:"maxPatients"`
	Registered  int    `json:"registered"`
}

type DeptSchedule struct {
	Date         string         `json:"date"`    // 2024-06-18, 2024-06-19, ...
	Weekday      int            `json:"weekday"` // 0:周日, 1:周一, ...
	Remain       int            `json:"remain"`
	DocScheduler []DocScheduler `json:"docScheduler"`
}

type ScheduleInfo struct {
	HosName      string         `json:"hosName"`
	FatherName   string         `json:"fatherName"`
	Name         string         `json:"name"`
	Total        int            `json:"total"`
	DeptSchedule []DeptSchedule `json:"deptSchedule"`
}

func (xh *XytHospitalHandler) docSchedules(ctx *gin.Context) {
	hosId := ctx.Query("hosId")
	deptId := ctx.Query("deptId")

	if hosId == "" || deptId == "" {
		ctx.JSON(200, app.ErrBadRequestQuery)
		return
	}

	var hos xytmodel.Hospital
	err := xh.db.Model(&xytmodel.Hospital{}).Where("uid = ?", hosId).Take(&hos).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	var dept xytmodel.Department
	err = xh.db.Model(&xytmodel.Department{}).Where("uid = ?", deptId).Take(&dept).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	var deptOne xytmodel.Department
	err = xh.db.Model(&xytmodel.Department{}).Where("uid = ?", dept.ParentID).Take(&deptOne).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	var docs []xytmodel.Doctor
	err = xh.db.Model(&xytmodel.Doctor{}).Where("hos_id = ? and dept_id = ?", hosId, deptId).Find(&docs).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	if len(docs) <= 0 {
		ctx.JSON(200, app.ResponseOK(ScheduleInfo{}))
		return
	}

	pageNo, _ := strconv.Atoi(ctx.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "1"))
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	if pageSize > MaxSchedulerDays {
		pageSize = MaxSchedulerDays
	}

	offset := (pageNo - 1) * pageSize
	if offset > int(MaxSchedulerDays) {
		ctx.JSON(200, app.ErrOutOfRange)
		return
	}

	var endIndex = 0
	if pageNo*pageSize > MaxSchedulerDays {
		endIndex = MaxSchedulerDays
	} else {
		endIndex = pageNo * pageSize
	}

	var results = ScheduleInfo{
		HosName:    hos.FullName,
		FatherName: deptOne.Name,
		Name:       dept.Name,
		Total:      MaxSchedulerDays,
	}

	today := time.Now()
	for i := offset; i < endIndex; i++ {
		date := today.AddDate(0, 0, i)
		var result DeptSchedule
		result, err = xh.getOrCreateDocScheduler(hosId, deptId, date, docs)
		if err != nil {
			fmt.Println("err: ", err)
			ctx.JSON(200, app.ErrInternalServer)
			return
		}
		results.DeptSchedule = append(results.DeptSchedule, result)
	}
	ctx.JSON(200, app.ResponseOK(results))
}

func (xh *XytHospitalHandler) getOrCreateDocScheduler(hosId, deptId string, t time.Time, docs []xytmodel.Doctor) (DeptSchedule, error) {
	year, month, day := t.Date()
	date := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	var scheduler []xytmodel.Schedule
	err := xh.db.Model(&xytmodel.Schedule{}).Where("hos_id = ? and dept_id = ? and work_date = ?", hosId, deptId, date).Find(&scheduler).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return DeptSchedule{}, err
	}
	var amount = 0
	if len(scheduler) <= 0 {
		var result []xytmodel.Schedule
		for i := 0; i < len(docs); i++ {
			amount = 0
			switch docs[i].Rank {
			case "主任医师":
				amount = 30
			case "副主任医师":
				amount = 20
			case "特需门诊":
				amount = 100
			default:
				amount = 10
			}
			var sche = xytmodel.Schedule{
				ScheId:      utils.GenerateUinqueID(),
				DocId:       docs[i].Id,
				HosID:       hosId,
				DeptID:      deptId,
				WorkDate:    date,
				WorkWeek:    int(t.Weekday()),
				TimeSlot:    randomTimeSlot(),
				Amount:      amount,
				MaxPatients: MaxPatientsPerDay,
				Registered:  0,
			}
			err = xh.db.Model(&xytmodel.Schedule{}).Create(&sche).Error
			if err != nil {
				return DeptSchedule{}, err
			}
			result = append(result, sche)
		}
		return docSchedulerToView(result, docs), nil
	}
	return docSchedulerToView(scheduler, docs), nil
}

type AddPatientReq struct {
	UserId   string `json:"userId"`
	Name     string `json:"name"`
	IdNumber string `json:"idNumber"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
	Sex      bool   `json:"sex"`
}

type AddOrderReq struct {
	PatientId string `json:"patientId"`
	ScheId    string `json:"scheId"`
}

func (xh *XytHospitalHandler) addOrder(ctx *gin.Context) {
	var req AddOrderReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	orderId, err := CreateOrder(xh.db, req)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, app.ErrUserNotFound) {
			ctx.JSON(http.StatusOK, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseErr(404, app.ErrUserNotFound.Error()))
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(map[string]string{"orderId": orderId}))
}

func (xh *XytHospitalHandler) getOrder(ctx *gin.Context) {
	orderId := ctx.Query("orderId")
	if orderId == "" {
		ctx.JSON(200, app.ErrBadRequestQuery)
		return
	}

	var order xytmodel.RegisterOrder
	err := xh.db.Table(xytmodel.TableOrder).Where("order_id = ?", orderId).Take(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, app.ErrNotFound)
			return
		}
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(order))
}

func (xh *XytHospitalHandler) listOrder(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}
	queries := ctx.Request.URL.Query()
	dbQuery := xh.db.Model(&xytmodel.RegisterOrder{})
	for k, v := range queries {
		switch k {
		case "patient_id":
			if len(v) > 0 && v[0] != "" {
				dbQuery.Where("patient_id = ?", v[0])
			}
		case "state":
			if len(v) > 0 && v[0] != "" {
				st, err := strconv.Atoi(v[0])
				if err == nil {
					if st >= -1 && st <= 2 {
						dbQuery.Where("state = ?", v[0])
					}
				}
			}
		case "pageNo", "pageSize", "timeStamp":
			continue
		}
	}
	dbQuery.Where("user_id = ?", userid.(string))

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
		ctx.JSON(http.StatusOK, app.ResponsePageData(0, nil))
		return
	}

	var orderList []xytmodel.RegisterOrder
	err = dbQuery.Find(&orderList).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponsePageData(total, orderList))
}

type cancelOrderReq struct {
	OrderId string `json:"orderId"`
}

func (xh *XytHospitalHandler) cancelOrder(ctx *gin.Context) {
	var req cancelOrderReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	var order xytmodel.RegisterOrder
	err := xh.db.Table(xytmodel.TableOrder).Where("order_id = ?", req.OrderId).Take(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(200, app.ErrNotFound)
			return
		}
		ctx.JSON(200, app.ErrInternalServer)
		return
	}

	var sche xytmodel.Schedule
	err = xh.db.Table(xytmodel.TableSchedule).Where("sche_id = ?", order.ScheId).Take(&sche).Error
	if err != nil {
		ctx.JSON(200, app.ErrNotFound)
		return
	}

	switch order.State {
	case 0:
		err = xh.db.Table(xytmodel.TableOrder).Where("order_id = ?", req.OrderId).Update("state", -1).Error
		if err != nil {
			ctx.JSON(200, app.ErrInternalServer)
			return
		}
		if sche.Registered > 0 {
			sche.Registered = sche.Registered - 1
		}
		err = xh.db.Model(&sche).Update("registered", sche.Registered).Error
		if err != nil {
			ctx.JSON(200, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseOK(nil))
	case 1, 2:
		ctx.JSON(http.StatusOK, app.ResponseErr(403, "无法取消该订单"))
	default:
		ctx.JSON(http.StatusOK, app.ResponseOK(nil))
	}
}

type DocRegister struct {
	DocId      string `json:"docId"`
	DoctorName string `json:"doctorName"`
	Rank       string `json:"rank"`
	Profile    string `json:"profile"`
	WorkDay    string `json:"workDay"`
	HosName    string `json:"hosName"`
	DeptName   string `json:"deptName"`
	Amount     int    `json:"amount"`
}

func (xh *XytHospitalHandler) getDoctor(ctx *gin.Context) {
	scheId := ctx.Query("scheId")

	var schedule xytmodel.Schedule
	err := xh.db.Table(xytmodel.TableSchedule).Where("sche_id = ?", scheId).Take(&schedule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, app.ErrNotFound)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	var doctor xytmodel.Doctor
	err = xh.db.Table(xytmodel.TableDoctor).Where("id = ?", schedule.DocId).Take(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, app.ErrNotFound)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	var hos xytmodel.Hospital
	err = xh.db.Table(xytmodel.TableHospital).Where("uid = ?", schedule.HosID).Take(&hos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, app.ErrNotFound)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	var dept xytmodel.Department
	err = xh.db.Table(xytmodel.TableDepartment).Where("uid = ? and hospital_id = ?", schedule.DeptID, schedule.HosID).Take(&dept).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, app.ErrNotFound)
			return
		}
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	var amount = 0
	switch doctor.Rank {
	case "主任医师":
		amount = 30
	case "副主任医师":
		amount = 20
	case "特需门诊":
		amount = 100
	default:
		amount = 10
	}

	var result = DocRegister{
		DocId:      schedule.DocId,
		DoctorName: doctor.Name,
		Rank:       doctor.Rank,
		Profile:    doctor.Profile,
		WorkDay:    schedule.WorkDate,
		HosName:    hos.FullName,
		DeptName:   dept.Name,
		Amount:     amount,
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(result))
}

func docSchedulerToView(sche []xytmodel.Schedule, docs []xytmodel.Doctor) DeptSchedule {
	var result = DeptSchedule{
		Date:    sche[0].WorkDate,
		Weekday: sche[0].WorkWeek,
	}
	var remain = 0
	for i := 0; i < len(sche); i++ {
		result.DocScheduler = append(result.DocScheduler, DocScheduler{
			DocId:       sche[i].DocId,
			ScheId:      sche[i].ScheId,
			TimeSlot:    sche[i].TimeSlot,
			DoctorName:  docs[i].Name,
			Rank:        docs[i].Rank,
			Profile:     docs[i].Profile,
			WorkDay:     sche[i].WorkDate,
			Amount:      sche[i].Amount,
			MaxPatients: sche[i].MaxPatients,
			Registered:  sche[i].Registered,
		})
		remain += sche[i].MaxPatients - sche[i].Registered
	}
	result.Remain = remain
	return result
}

type DepartmentResponse struct {
	UID         string               `json:"uid"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Children    []DepartmentResponse `json:"children,omitempty"`
}

func BuildDepartmentTree(depts []xytmodel.Department) []DepartmentResponse {
	// 先构建 map：parentID -> []children
	childrenMap := make(map[string][]DepartmentResponse)
	var roots []DepartmentResponse

	for _, dept := range depts {
		node := DepartmentResponse{
			UID:         dept.UID,
			Name:        dept.Name,
			Description: dept.Description,
		}
		childrenMap[dept.ParentID] = append(childrenMap[dept.ParentID], node)
	}

	// 构建最终树形
	for _, root := range childrenMap[""] { // 顶级科室
		fillChildren(&root, childrenMap)
		roots = append(roots, root)
	}

	return roots
}

func fillChildren(node *DepartmentResponse, childrenMap map[string][]DepartmentResponse) {
	children := childrenMap[node.UID]
	for i := range children {
		fillChildren(&children[i], childrenMap)
	}
	node.Children = children
}

func randomTimeSlot() string {
	surnames := []string{
		"上午",
		"下午",
	}
	rand.NewSource(time.Now().UnixNano())
	return surnames[rand.Intn(len(surnames))]
}

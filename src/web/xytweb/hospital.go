package xytweb

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/types/app"
	"gorm.io/gorm"
)

type XytHospitalHandler struct {
	db *gorm.DB
}

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
}

func (xh *XytHospitalHandler) hosList(ctx *gin.Context) {
	var err error
	var hoslist []xytmodel.Hospital
	queries := ctx.Request.URL.Query()
	dbQuery := xh.db.Model(&xytmodel.Hospital{})
	for k, v := range queries {
		switch k {
		case "uid":
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
		err = xh.db.Table(xytmodel.TableCity).Where("city_name = ?", cityName).Take(&city).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(200, app.ResponseErr(404, "请指定一个存在的城市名字"))
				return
			}
			ctx.JSON(200, app.ErrInternalServer)
			return
		}
		city_code = city.CityCode
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
	uid := ctx.Query("uid")
	if uid == "" {
		ctx.JSON(200, app.ResponseErr(400, "请指定医院uid"))
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

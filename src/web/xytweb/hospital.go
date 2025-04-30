package xytweb

import (
	"errors"
	"fmt"
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
	// ug := group.Group("/hos")
	group.GET("/hos/list", xh.hoslist)
	group.GET("/hos/grade", xh.hosgrade)
	group.GET("/hos/region", xh.hosregion)
}

func (xh *XytHospitalHandler) hoslist(ctx *gin.Context) {
	queries := ctx.Request.URL.Query()
	fmt.Println(queries)
	dbQuery := xh.db.Debug().Model(&xytmodel.Hospital{})
	for k, v := range queries {
		switch k {
		case "uid":
			dbQuery.Where("uid = ?", v[0])
		case "gradeName":
			dbQuery.Where("grade_name = ?", v[0])
		case "cityCode":
			dbQuery.Where("city_code = ?", v[0])
		case "cityName":
			dbQuery.Where("city_name = ?", v[0])
		case "districtName":
			dbQuery.Where("district_name = ?", v[0])
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
		ctx.JSON(200, app.ResponsePageData(total, nil))
		return
	}
	fmt.Println(total)

	if total < 1 {
		ctx.JSON(http.StatusOK, app.ResponsePageData(0, nil))
		return
	}

	var hoslist []xytmodel.Hospital
	err = dbQuery.Limit(pageSize).Offset(offset).Find(&hoslist).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponsePageData(total, hoslist))
}

func (xh *XytHospitalHandler) hosgrade(ctx *gin.Context) {
	dbQuery := xh.db.Table(xytmodel.TableHospitalGrade)
	var hosgrade []xytmodel.HospitalGrade
	err := dbQuery.Find(&hosgrade).Error
	if err != nil {
		ctx.JSON(200, app.ErrInternalServer)
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseOK(hosgrade))
}

func (xh *XytHospitalHandler) hosregion(ctx *gin.Context) {
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
				ctx.JSON(200, app.ResponseErr(400, "请指定一个存在的城市名字"))
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

func buildWhere(db *gorm.DB, k string, v []string) {
	length := len(v)
	unescapeKey, err := url.QueryUnescape(k)
	if err == nil {
		k = unescapeKey
	}
	switch length {
	case 0:
		return
	case 1:
		db.Where(fmt.Sprintf("%s = ?", k), v[0])
		return
	default:
		orGroup := db.Session(&gorm.Session{DryRun: true})
		for _, val := range v {
			orGroup = orGroup.Or(fmt.Sprintf("%s = ?", k), val)
		}
		db.Where(orGroup)
		return
	}
}

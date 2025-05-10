package xytweb

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/types/app"
	"gorm.io/gorm"
)

type XytCitesHandler struct {
	db *gorm.DB
}

func NewXytCiteslHandler(db *gorm.DB) *XytCitesHandler {
	return &XytCitesHandler{
		db: db,
	}
}

func (xh *XytCitesHandler) RegisterRoutes(group *gin.RouterGroup) {
	// ---------------- vbook api ---------------------
	ug := group.Group("/city")
	ug.GET("/cascader", xh.getCascader)
}

type CascaderCityData struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Leaf bool   `json:"leaf"`
}

func (xh *XytCitesHandler) getCascader(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusOK, app.ErrBadRequestQuery)
		return
	}
	var err error
	switch len(code) {
	case 2:
		if code == "86" {
			var province []xytmodel.Province
			err = xh.db.Table(xytmodel.TableProvince).Find(&province).Error
			if err != nil {
				ctx.JSON(http.StatusOK, app.ErrInternalServer)
				return
			}
			ctx.JSON(http.StatusOK, app.ResponseOK(provinceToCascaderCityData(province)))
			return
		}
		var city []xytmodel.City
		err = xh.db.Table(xytmodel.TableCity).Where("province_code = ?", code).Find(&city).Error
		if err != nil {
			ctx.JSON(http.StatusOK, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseOK(cityToCascaderCityData(city)))
	case 4:
		var district []xytmodel.District
		err = xh.db.Table(xytmodel.TableDistrict).Where("city_code = ?", code).Find(&district).Error
		if err != nil {
			ctx.JSON(http.StatusOK, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseOK(districtToCascaderCityData(district)))
	default:
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
	}
}

func provinceToCascaderCityData(province []xytmodel.Province) []CascaderCityData {
	var result = make([]CascaderCityData, len(province))
	for k, v := range province {
		result[k] = CascaderCityData{
			Code: v.Code,
			Name: v.Name,
			Leaf: true,
		}
	}
	return result
}

func cityToCascaderCityData(city []xytmodel.City) []CascaderCityData {
	var result = make([]CascaderCityData, len(city))
	for k, v := range city {
		result[k] = CascaderCityData{
			Code: v.Code,
			Name: v.Name,
			Leaf: true,
		}
	}
	return result
}

func districtToCascaderCityData(district []xytmodel.District) []CascaderCityData {
	var result = make([]CascaderCityData, len(district))
	for k, v := range district {
		result[k] = CascaderCityData{
			Code: v.Code,
			Name: v.Name,
			Leaf: false,
		}
	}
	return result
}

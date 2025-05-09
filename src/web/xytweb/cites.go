package xytweb

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/config"
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
	ug := group.Group("/cites")
	ug.GET("/province", xh.getPrivince)
}

func (xh *XytCitesHandler) getPrivince(ctx *gin.Context) {
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

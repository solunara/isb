package xytweb

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
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
		Name: xytuser.Nickname,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(map[string]string{"name": xytuser.Nickname, "token": tokenVal}))
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
		xytuser.Nickname = fmt.Sprintf("用户_%s****%s", phone[:3], phone[len(phone)-4:])
		xytuser.Phone = sql.NullString{
			String: phone,
			Valid:  true,
		}
		tnow := time.Now().Unix()
		xytuser.Ctime = tnow
		xytuser.Utime = tnow
		err = db.Create(&xytuser).Error
		if err != nil {
			return xytmodel.XytUser{}, err
		}
		return xytuser, nil
	}
	return xytuser, nil
}

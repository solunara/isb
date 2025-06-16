package hllweb

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/config"
	"github.com/solunara/isb/src/model/hllmodel"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type HllUserHandler struct {
	cache redis.Cmdable
	db    *gorm.DB
}

func NewHllUserlHandler(cache redis.Cmdable, db *gorm.DB) *HllUserHandler {
	return &HllUserHandler{
		cache: cache,
		db:    db,
	}
}

func (h *HllUserHandler) RegisterRoutes(group *gin.RouterGroup) {
	// ---------------- hllmgr api ---------------------
	ug := group.Group("/user")
	ug.POST("/login", h.LoginWithPwd)
	ug.GET("/info", h.GetUserInfo)
}

func (h *HllUserHandler) GetUserInfo(ctx *gin.Context) {
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}
	fmt.Println("userid: ", userid)
	u, err := FindUserById(h.db, userid.(string))
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, app.ResponseOK(map[string]string{"name": u.Username}))
	case app.ErrInvalidUserOrPassword:
		ctx.JSON(http.StatusOK, app.ErrBadRequestErrInvalidUserOrPassword)
	default:
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
	}
}

func FindUserById(db *gorm.DB, userid string) (hllmodel.HllUser, error) {
	var hlluser hllmodel.HllUser
	err := db.Table(hllmodel.TableHllUser).Where("user_id = ?", userid).Take(&hlluser).Error
	if err != nil {
		return hllmodel.HllUser{}, err
	}
	return hlluser, nil
}

func (h *HllUserHandler) LoginWithPwd(ctx *gin.Context) {
	type Req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		ctx.JSON(http.StatusOK, app.ErrEmptyRequest)
		return
	}

	u, err := FindOrCreateUser(h.db, req.Username, req.Password)
	switch err {
	case nil:
		tokenVal, err := jwtoken.NewJWToken().CreateJWToken(jwtoken.CustomClaims{
			Name:   u.Username,
			UserId: u.UserId,
		})
		if err != nil {
			ctx.JSON(http.StatusOK, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseOK(map[string]string{"token": tokenVal}))
	case app.ErrInvalidUserOrPassword:
		ctx.JSON(http.StatusOK, app.ErrBadRequestErrInvalidUserOrPassword)
	default:
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
	}
}

func FindOrCreateUser(db *gorm.DB, username, password string) (hllmodel.HllUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return hllmodel.HllUser{}, err
	}

	var hlluser hllmodel.HllUser
	err = db.Table(hllmodel.TableHllUser).Where("username = ?", username).Take(&hlluser).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return hllmodel.HllUser{}, err
		}
		hlluser.UserId = uuid.New().String()
		hlluser.Username = username

		hlluser.Password = string(hash)
		hlluser.State = "active"
		err = db.Create(&hlluser).Error
		if err != nil {
			return hllmodel.HllUser{}, err
		}
		return hlluser, nil
	}

	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(hlluser.Password), []byte(password))
	if err != nil {
		return hllmodel.HllUser{}, app.ErrInvalidUserOrPassword
	}

	return hlluser, nil
}

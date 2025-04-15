package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/service"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
)

type MsUserHandler struct {
	usersvc service.MsUserService
}

func NewMsUserHandler(usersvc service.MsUserService) *MsUserHandler {
	return &MsUserHandler{
		usersvc: usersvc,
	}
}

func (h *MsUserHandler) RegisterRoutes(group *gin.RouterGroup) {
	// ---------------- msfs api ---------------------
	group.POST("/signup", h.SignUp)
	group.POST("/login", h.LoginWithPwd)
}

func (h *MsUserHandler) SignUp(ctx *gin.Context) {
	// type SignUpReq struct {
	// 	Password        string `json:"password"`
	// 	ConfirmPassword string `json:"confirm_password"`
	// }

	// var req SignUpReq
	// if err := ctx.Bind(&req); err != nil {
	// 	ctx.JSON(http.StatusOK, app.ErrBadRequest)
	// 	return
	// }

	// isEmail, err := h.emailRexExp.MatchString(req.Email)
	// if err != nil {
	// 	ctx.JSON(http.StatusOK, app.ErrInternalServer)
	// 	return
	// }

	// if !isEmail {
	// 	ctx.JSON(http.StatusOK, app.ErrBadRequestInvalidEmail)
	// 	return
	// }

	// if req.Password != req.ConfirmPassword {
	// 	ctx.JSON(http.StatusOK, app.ErrBadRequestWrongPassword)
	// 	return
	// }

	// isPassword, err := h.passwordRexExp.MatchString(req.Password)
	// if err != nil {
	// 	ctx.JSON(http.StatusOK, app.ErrInternalServer)
	// 	return
	// }

	// if !isPassword {
	// 	ctx.JSON(http.StatusOK, app.ErrBadRequestInvalidPassword)
	// 	return
	// }

	// err = h.usersvc.Signup(ctx, repository.User{
	// 	Email:    req.Username,
	// 	Password: req.Password,
	// })

	// switch err {
	// case nil:
	// 	ctx.JSON(http.StatusOK, app.ResponseOK("registration successful"))
	// case app.ErrDuplicateEmail:
	// 	ctx.JSON(http.StatusOK, app.ResponseErr(400, app.ErrDuplicateEmail.Error()))
	// default:
	// 	ctx.JSON(http.StatusOK, app.ErrInternalServer)
	// }
}

func (h *MsUserHandler) LoginWithPwd(ctx *gin.Context) {
	type Req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}
	u, err := h.usersvc.LoginWithPwd(ctx, req.Username, req.Password)
	switch err {
	case nil:
		tokenVal, err := jwtoken.NewJWToken().CreateJWToken(jwtoken.CustomClaims{
			Name: u.Username,
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

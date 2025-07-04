package web

import (
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/repository/cache"
	"github.com/solunara/isb/src/service"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
)

const biz_login = "user_login"

// 确保 UserHandler 实现了 handler 接口
var _ handler = &UserHandler{}

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	codeSvc        service.CaptchaService
	usersvc        service.UserService
}

func NewUserHandler(usersvc service.UserService, codeSvc service.CaptchaService) *UserHandler {
	const (
		// 邮箱格式校验
		emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		// 密码格式校验
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	)
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		usersvc:        usersvc,
		codeSvc:        codeSvc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/user")
	ug.POST("/signup", h.SignUp)
	ug.POST("/logout", h.Logout)

	ug.POST("/login/email", h.LoginWithEmail)
	ug.POST("/login/sms/send", h.LoginSMSSend)
	ug.POST("/login/sms", h.LoginSMS)

	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	if !isEmail {
		ctx.JSON(http.StatusOK, app.ErrBadRequestInvalidEmail)
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, app.ErrBadRequestWrongPassword)
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	if !isPassword {
		ctx.JSON(http.StatusOK, app.ErrBadRequestInvalidPassword)
		return
	}

	err = h.usersvc.Signup(ctx, repository.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		ctx.JSON(http.StatusOK, app.ResponseOK("registration successful"))
	case app.ErrDuplicateEmail:
		ctx.JSON(http.StatusOK, app.ResponseErr(400, app.ErrDuplicateEmail.Error()))
	default:
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
	}
}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		//Secure: true,
		//HttpOnly: true,
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出登录成功")
}

func (h *UserHandler) LoginWithEmail(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}
	u, err := h.usersvc.LoginWithEmailPwd(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十分钟
			MaxAge: 600,
			//Secure:   true,
			HttpOnly: true,
		})
		err = sess.Save()
		if err != nil {
			ctx.JSON(http.StatusOK, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseOK("登录成功"))
	case app.ErrInvalidUserOrPassword:
		ctx.JSON(http.StatusOK, app.ErrBadRequestErrInvalidUserOrPassword)
	default:
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
	}
}

func (h *UserHandler) LoginSMSSend(ctx *gin.Context) {
	type LoginSMSSendCodeReq struct {
		Phone string `json:"phone"`
	}

	var req LoginSMSSendCodeReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	if req.Phone == "" {
		ctx.JSON(http.StatusOK, app.ErrEmptyRequest)
		return
	}

	err := h.codeSvc.Send(ctx, biz_login, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, app.ResponseOK(nil))
	case cache.ErrSendTooFrequently:
		ctx.JSON(http.StatusOK, app.ResponseErr(400, "sent too often"))
	default:
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type LoginSMSVerifyCodeReq struct {
		Phone   string `json:"phone"`
		Captcha string `json:"captcha"`
	}

	var req LoginSMSVerifyCodeReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	ok, err := u.codeSvc.Verify(ctx, biz_login, req.Phone, req.Captcha)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrBadPhoneOrCode)
		return
	}

	user, err := u.usersvc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	tokenVal, err := jwtoken.NewJWToken().CreateJWToken(jwtoken.CustomClaims{
		Name: user.Nickname,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}

	ctx.JSON(http.StatusOK, app.ResponseOK(map[string]string{"name": user.Nickname, "token": tokenVal}))
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Nickname string `json:"nickname"`
		Profile  string `json:"profile"`
		Birthday string `json:"birthday"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequest)
		return
	}

	birthdaytime, err := time.Parse(time.DateTime, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrBadRequestWrongBirthday)
		return
	}

	u, err := h.usersvc.EditProfile(ctx, repository.User{
		Id:       ctx.GetInt64("userId"),
		Nickname: req.Nickname,
		Profile:  req.Profile,
		Birthday: birthdaytime,
	})

	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十分钟
			MaxAge: 600,
			//Secure:   true,
			HttpOnly: true,
		})
		err = sess.Save()
		if err != nil {
			ctx.JSON(http.StatusOK, app.ErrInternalServer)
			return
		}
		ctx.JSON(http.StatusOK, app.ResponseOK("修改成功"))
	case app.ErrInvalidUserOrPassword:
		ctx.JSON(http.StatusOK, app.ErrBadRequestErrInvalidUserOrPassword)
	default:
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
	}
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	//TODO
	ctx.JSON(http.StatusOK, app.ResponseOK(repository.User{}))
}

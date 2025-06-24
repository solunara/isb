package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/solunara/isb/src/service"
	"github.com/solunara/isb/src/service/oauth2/wechat"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	//jwtHandler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
		})
		return
	}
	if err = h.setStateCookie(ctx, state); err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseType{
		Code: 200,
		Msg:  "",
		Data: url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}

	code := ctx.Query("code")
	info, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 这里怎么办？
	// 从 userService 里面拿 uid
	u, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	tokenVal, err := jwtoken.NewJWToken().CreateJWToken(jwtoken.CustomClaims{
		Name: u.Nickname,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, app.ErrInternalServer)
		return
	}
	ctx.Header("x-jwt-token", tokenVal)
	ctx.JSON(http.StatusOK, app.ResponseOK(nil))
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie("x-cookie")
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return []byte("XJvJWOadrbgUxUwqcXOKnnGpVwWPKgCA"), nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context,
	state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("XJvJWOadrbgUxUwqcXOKnnGpVwWPKgCA"))
	if err != nil {

		return err
	}
	ctx.SetCookie("x-cookie", tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

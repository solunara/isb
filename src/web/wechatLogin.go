package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	url, err := h.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
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
	code := ctx.Query("code")
	state := ctx.Query("state")
	info, err := h.svc.VerifyCode(ctx, code, state)
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

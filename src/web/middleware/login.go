package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/config"
	"github.com/solunara/isb/src/types/app"
	"github.com/solunara/isb/src/types/jwtoken"
)

// LoginJWTMiddlewareBuilder JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		flag := false
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				flag = true
				break
			}
		}
		if !flag {
			token := ctx.GetHeader(config.HTTTP_HEADER_AUTH)
			if token == "" {
				ctx.AbortWithStatusJSON(200, app.ErrForbidden)
				return
			}
			fmt.Println("token: ", token)
			claims, err := jwtoken.NewJWToken().ParesJWToken(token)
			if err != nil || claims == nil {
				ctx.AbortWithStatusJSON(200, app.ErrForbidden)
				return
			}
			fmt.Println("token id: ", claims.UserId)
			ctx.Set(config.USER_ID, claims.UserId)
		}
		ctx.Next()
	}
}

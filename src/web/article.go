package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solunara/isb/src/config"
	"github.com/solunara/isb/src/service"
	"github.com/solunara/isb/src/types/app"
)

const biz_article = "article"

// 确保 UserHandler 实现了 handler 接口
var _ handler = &UserHandler{}

type ArticleHandler struct {
	svc service.ArticleService
}

func NewArticleHandler(svc service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	// 修改/新增
	g.POST("/edit", h.Edit)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	userid, ok := ctx.Get(config.USER_ID)
	if !ok {
		ctx.JSON(http.StatusOK, app.ErrUnauthorized)
		return
	}

	id, err := h.svc.Save(ctx, req.toDomain(userid.(int64)))
	if err != nil {
		ctx.JSON(http.StatusOK, app.ResponseType{
			Code: 5,
			Msg:  "系统错误",
		})
		// 打日志
		return
	}
	ctx.JSON(http.StatusOK, app.ResponseType{
		Data: id,
	})
}

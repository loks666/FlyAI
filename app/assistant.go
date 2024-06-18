package app

import (
	"github.com/gin-gonic/gin"
)

// SaveAssistant
// @description: 保存AI助手
// @param ctx
func SaveAssistant(ctx *gin.Context) {

	//ctx.String(http.StatusOK, "操作成功")
	ctx.Redirect(302, "/assistant.html")
}

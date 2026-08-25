package app

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	newstask "github.com/loks666/FlyAI/tasks/news"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync"
)

type runAINewsParam struct {
	WxId   string `json:"wxId" binding:"required"`
	DryRun bool   `json:"dryRun"`
}

var aiNewsSendState sync.Map

// RunAINews previews or sends the current hot-list morning brief to one enabled group.
func RunAINews(ctx *gin.Context) {
	var p runAINewsParam
	if err := ctx.ShouldBindJSON(&p); err != nil || !strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "群参数错误"})
		return
	}
	group, err := service.GetFriendInfoById(p.WxId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "群不存在"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取群配置失败"})
		return
	}
	if !group.IsOk || !group.EnableMonitor || !group.EnableNews {
		ctx.JSON(http.StatusConflict, gin.H{"message": "请先开启该群的监控和每日热点早报"})
		return
	}
	if _, loaded := aiNewsSendState.LoadOrStore(p.WxId, true); loaded {
		ctx.JSON(http.StatusConflict, gin.H{"message": "该群每日热点早报正在生成"})
		return
	}
	defer aiNewsSendState.Delete(p.WxId)

	if p.DryRun {
		preview, buildErr := newstask.BuildMorningPost()
		if buildErr != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"message": buildErr.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "每日热点早报预览生成成功，未发送", "preview": preview})
		return
	}
	if _, err = newstask.SendMorningPostToGroup(p.WxId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "每日热点早报发送失败：" + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "每日热点早报已发送"})
}

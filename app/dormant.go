package app

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/service"
	dormanttask "github.com/loks666/FlyAI/tasks/dormant"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync"
)

var dormantSendState sync.Map

func ChangeEnableDormantStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil || !strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "群参数错误"})
		return
	}
	var record entity.Friend
	if err := client.MySQL.Select("wxid", "enable_dormant").Where("wxid = ?", p.WxId).First(&record).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "群聊不存在"})
		return
	}
	enabled := !record.EnableDormant
	if err := client.MySQL.Model(&entity.Friend{}).Where("wxid = ?", p.WxId).Update("enable_dormant", enabled).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "修改潜水统计状态失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

type runDormantParam struct {
	WxId   string `json:"wxId" binding:"required"`
	DryRun bool   `json:"dryRun"`
}

func RunDormantReport(ctx *gin.Context) {
	var p runDormantParam
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
	if !group.IsOk || !group.EnableMonitor || !group.EnableDormant {
		ctx.JSON(http.StatusConflict, gin.H{"message": "请先开启该群的监控和潜水成员统计"})
		return
	}
	if _, loaded := dormantSendState.LoadOrStore(p.WxId, true); loaded {
		ctx.JSON(http.StatusConflict, gin.H{"message": "该群潜水成员统计正在生成"})
		return
	}
	defer dormantSendState.Delete(p.WxId)

	if p.DryRun {
		preview, buildErr := dormanttask.BuildReport(p.WxId)
		if buildErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": buildErr.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "潜水成员统计预览生成成功，未发送", "preview": preview})
		return
	}
	if _, err = dormanttask.SendReportToGroup(p.WxId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "潜水成员统计发送失败：" + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "潜水成员统计已发送"})
}

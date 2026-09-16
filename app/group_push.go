package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	dormanttask "github.com/loks666/FlyAI/tasks/dormant"
	newstask "github.com/loks666/FlyAI/tasks/news"
	"github.com/loks666/FlyAI/tasks/watergroup"
	"gorm.io/gorm"
)

type runGroupFeatureParam struct {
	WxId    string `json:"wxId" binding:"required"`
	Feature string `json:"feature" binding:"required"`
}

type groupFeaturePusher struct {
	label string
	run   func(string) (string, error)
}

var groupFeaturePushState sync.Map

func findGroupFeaturePusher(feature string) (groupFeaturePusher, bool) {
	pushers := map[string]groupFeaturePusher{
		service.FeatureWeekly:  {label: "排行榜", run: watergroup.SendWeekToGroup},
		service.FeatureYearly:  {label: "年排行榜", run: watergroup.SendYearToGroup},
		service.FeatureNews:    {label: "热点早报", run: newstask.SendMorningPostToGroup},
		service.FeatureDormant: {label: "潜水统计", run: dormanttask.SendReportToGroup},
	}
	pusher, ok := pushers[feature]
	return pusher, ok
}

// RunGroupFeature manually sends one selected feature to one current group.
// Manual delivery is intentionally independent from the automatic schedule switch.
func RunGroupFeature(ctx *gin.Context) {
	var p runGroupFeatureParam
	if err := ctx.ShouldBindJSON(&p); err != nil || !strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "群聊或推送参数错误"})
		return
	}

	pusher, ok := findGroupFeaturePusher(p.Feature)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "未知群聊推送功能"})
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
	if !group.IsOk {
		ctx.JSON(http.StatusConflict, gin.H{"message": "该群已不在当前微信通讯录中"})
		return
	}

	stateKey := p.Feature + ":" + p.WxId
	if _, loaded := groupFeaturePushState.LoadOrStore(stateKey, true); loaded {
		ctx.JSON(http.StatusConflict, gin.H{"message": pusher.label + "正在生成，请勿重复点击"})
		return
	}
	defer groupFeaturePushState.Delete(stateKey)

	if _, err = pusher.run(p.WxId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("%s发送失败：%v", pusher.label, err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": pusher.label + "已发送"})
}

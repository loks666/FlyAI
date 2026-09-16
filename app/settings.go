package app

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	"net/http"
)

type updateFeatureTimeParam struct {
	Feature          string `json:"feature" binding:"required"`
	Schedule         string `json:"schedule" binding:"required"`
	SendTime         string `json:"sendTime" binding:"required"`
	MessageThreshold *int   `json:"messageThreshold"`
}

func GetFeatureSettings(ctx *gin.Context) {
	newsSetting, newsErr := service.GetFeatureSetting(service.FeatureNews)
	weeklySetting, weeklyErr := service.GetFeatureSetting(service.FeatureWeekly)
	yearlySetting, yearlyErr := service.GetFeatureSetting(service.FeatureYearly)
	dormantSetting, dormantErr := service.GetFeatureSetting(service.FeatureDormant)
	if newsErr != nil || weeklyErr != nil || yearlyErr != nil || dormantErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取功能设置失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"news":    newsSetting,
		"weekly":  weeklySetting,
		"yearly":  yearlySetting,
		"dormant": dormantSetting,
	})
}

func UpdateFeatureSettingTime(ctx *gin.Context) {
	var p updateFeatureTimeParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}
	setting, err := service.UpdateFeatureSetting(p.Feature, p.Schedule, p.SendTime, p.MessageThreshold)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "推送计划已保存并立即生效", "setting": setting})
}

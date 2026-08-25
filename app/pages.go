package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/service"
	"net/http"
)

// Index
// @description: 首页
// @param ctx
func Index(ctx *gin.Context) {
	var result = gin.H{
		"msg": "success",
	}
	service.SyncFriendCatalog()
	// 取出群组列表
	_, groups, err := service.GetAllFriend()
	if err != nil {
		result["msg"] = fmt.Sprintf("数据获取失败: %s", err.Error())
	}
	var activeGroupCount, monitored, ranked int
	for _, d := range groups {
		if !d.IsOk {
			continue
		}
		activeGroupCount++
		if d.EnableMonitor {
			monitored++
		}
		if d.EnableMonitor && d.EnableChatRank {
			ranked++
		}
	}
	result["groupCount"] = activeGroupCount
	result["monitorCount"] = monitored
	result["rankCount"] = ranked
	result["weeklyTime"] = featureSendTime(service.FeatureWeekly, "09:30")
	result["weeklyScheduleLabel"] = featureScheduleLabel(service.FeatureWeekly, service.ScheduleMonday)

	// 渲染页面
	ctx.HTML(http.StatusOK, "index.html", result)
}

// Friend
// @description: 好友列表
// @param ctx
func Friend(ctx *gin.Context) {
	var result = gin.H{
		"msg": "success",
	}
	service.SyncFriendCatalog()

	// 取出所有好友列表
	friends, _, err := service.GetAllFriend()
	if err != nil {
		result["msg"] = fmt.Sprintf("数据获取失败: %s", err.Error())
	}
	activeFriends := friends[:0]
	for _, friend := range friends {
		if friend.IsOk {
			activeFriends = append(activeFriends, friend)
		}
	}
	result["friends"] = activeFriends
	result["contactCount"] = len(activeFriends)
	// 渲染页面
	ctx.HTML(http.StatusOK, "friend.html", result)
}

// Group
// @description: 群组列表
// @param ctx
func Group(ctx *gin.Context) {
	var result = gin.H{
		"msg": "success",
	}
	service.SyncFriendCatalog()
	// 取出所有好友列表
	_, groups, err := service.GetAllFriend()
	if err != nil {
		result["msg"] = fmt.Sprintf("数据获取失败: %s", err.Error())
	}
	activeGroups := groups[:0]
	for _, group := range groups {
		if group.IsOk {
			activeGroups = append(activeGroups, group)
		}
	}
	result["groups"] = activeGroups
	result["newsTime"] = featureSendTime(service.FeatureNews, "08:00")
	result["weeklyTime"] = featureSendTime(service.FeatureWeekly, "09:30")
	result["dormantTime"] = featureSendTime(service.FeatureDormant, "09:00")
	result["dormantThreshold"] = featureMessageThreshold(service.FeatureDormant, 1)
	result["newsScheduleLabel"] = featureScheduleLabel(service.FeatureNews, service.ScheduleDaily)
	result["weeklyScheduleLabel"] = featureScheduleLabel(service.FeatureWeekly, service.ScheduleMonday)
	result["dormantScheduleLabel"] = featureScheduleLabel(service.FeatureDormant, service.ScheduleMonday)

	// 渲染页面
	ctx.HTML(http.StatusOK, "group.html", result)
}

// Settings renders the global schedule page.
func Settings(ctx *gin.Context) {
	var result = gin.H{"msg": "success"}
	newsSetting, newsErr := service.GetFeatureSetting(service.FeatureNews)
	weeklySetting, weeklyErr := service.GetFeatureSetting(service.FeatureWeekly)
	dormantSetting, dormantErr := service.GetFeatureSetting(service.FeatureDormant)
	if newsErr != nil || weeklyErr != nil || dormantErr != nil {
		result["msg"] = "读取统一功能设置失败"
	}
	owners, ownersErr := service.GetOwners()
	if ownersErr != nil {
		result["msg"] = "读取主人配置失败"
	}
	robotName := current.GetRobotInfo().Name
	if robotName == "" {
		robotName = "Fly AI"
	}
	result["owners"] = owners
	result["robotName"] = robotName
	result["newsTime"] = newsSetting.SendTime
	result["newsSchedule"] = newsSetting.Schedule
	result["weeklyTime"] = weeklySetting.SendTime
	result["weeklySchedule"] = weeklySetting.Schedule
	result["dormantTime"] = dormantSetting.SendTime
	result["dormantSchedule"] = dormantSetting.Schedule
	result["dormantThreshold"] = dormantSetting.MessageThreshold
	ctx.HTML(http.StatusOK, "settings.html", result)
}

func featureSendTime(feature, fallback string) string {
	setting, err := service.GetFeatureSetting(feature)
	if err != nil || setting.SendTime == "" {
		return fallback
	}
	return setting.SendTime
}

func featureScheduleLabel(feature, fallback string) string {
	setting, err := service.GetFeatureSetting(feature)
	if err != nil || setting.Schedule == "" {
		return service.FeatureScheduleLabel(fallback)
	}
	return service.FeatureScheduleLabel(setting.Schedule)
}

func featureMessageThreshold(feature string, fallback int) int {
	setting, err := service.GetFeatureSetting(feature)
	if err != nil || setting.MessageThreshold < 1 {
		return fallback
	}
	return setting.MessageThreshold
}

// Assistant
// @description: AI角色
// @param ctx
func Assistant(ctx *gin.Context) {
	var result = gin.H{
		"msg": "success",
	}

	result["aiModels"] = config.Conf.Ai.Models
	result["assistant"], _ = service.GetAllAiAssistant()

	// 渲染页面
	ctx.HTML(http.StatusOK, "assistant.html", result)
}

// ManageWithGroupAdmin
// @description: 群组管理（管理员可用）
// @param ctx
func ManageWithGroupAdmin(ctx *gin.Context) {
	// 取出id
	id := ctx.Query("id")
	if id == "" {
		ctx.HTML(http.StatusOK, "404.html", nil)
		return
	}

	var result = gin.H{
		"msg": "success",
	}
	result["info"], _ = service.GetFriendInfoById(id)
	result["aiModels"] = config.Conf.Ai.Models
	result["assistant"], _ = service.GetAllAiAssistant()
	// 渲染页面
	ctx.HTML(http.StatusOK, "manager-one.html", result)
}

// PageNotFound
// @description: 404页面
// @param ctx
func PageNotFound(ctx *gin.Context) {
	// 渲染页面
	ctx.HTML(http.StatusNotFound, "404.html", nil)
}

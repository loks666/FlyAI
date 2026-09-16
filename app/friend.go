package app

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/service"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
)

// changeStatusParam
// @description: 修改状态用的参数集
type changeStatusParam struct {
	WxId   string `json:"wxId" binding:"required"`
	UserId string `json:"userId"`
}

type changeGroupFeatureParam struct {
	WxId    string `json:"wxId" binding:"required"`
	Feature string `json:"feature" binding:"required"`
}

// ChangeGroupFeatureStatus toggles one automatic push for one group.
func ChangeGroupFeatureStatus(ctx *gin.Context) {
	var p changeGroupFeatureParam
	if err := ctx.ShouldBindJSON(&p); err != nil || !strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "群聊或功能参数错误"})
		return
	}
	var record entity.Friend
	if err := client.MySQL.Where("wxid = ?", p.WxId).First(&record).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "群聊不存在"})
		return
	}
	columns := map[string]struct {
		name    string
		enabled bool
	}{
		"weekly":  {name: "enable_weekly_rank", enabled: record.EnableWeeklyRank},
		"news":    {name: "enable_news", enabled: record.EnableNews},
		"dormant": {name: "enable_dormant", enabled: record.EnableDormant},
	}
	column, ok := columns[p.Feature]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "未知群聊推送功能"})
		return
	}
	enabled := !column.enabled
	if err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update(column.name, enabled).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "修改群聊推送状态失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"feature": p.Feature, "enabled": enabled})
}

// changeUseAiModelParam
// @description: 修改使用的AI模型用的参数集
type changeUseAiModelParam struct {
	WxId  string `json:"wxid" binding:"required"` // 群Id或微信Id
	Model string `json:"model"`                   // 模型代码
}

// autoClearMembers
// @description: 自动清理群成员
type autoClearMembers struct {
	WxId string `json:"wxid" binding:"required"` // 群Id
	Days int    `json:"days"`                    // 多少天未发言
}

// ChangeEnableMonitorStatus toggles the master switch for all group features.
func ChangeEnableMonitorStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	var record entity.Friend
	if err := client.MySQL.Select("wxid", "enable_monitor").
		Where("wxid = ?", p.WxId).
		Where("wxid LIKE '%@chatroom'").
		First(&record).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "群聊不存在"})
		return
	}

	enabled := !record.EnableMonitor
	if err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("enable_monitor", enabled).Error; err != nil {
		log.Printf("修改群监控状态失败：%s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

// GetEnableMonitorGroups is consumed by the local WeChat database bridge.
// It returns all current contact groups so manual owner commands retain live
// history even when automatic monitoring and scheduled delivery are disabled.
func GetEnableMonitorGroups(ctx *gin.Context) {
	records, err := service.GetAllBridgeGroups()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "获取桥接群失败"})
		return
	}
	groups := make([]string, 0, len(records))
	for _, record := range records {
		groups = append(groups, record.Wxid)
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups})
}

// ChangeEnableAiStatus
// @description: 修改是否开启AI
// @param ctx
func ChangeEnableAiStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的微信Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_ai`", gorm.Expr(" !`enable_ai`")).Error
	if err != nil {
		log.Printf("修改是否开启AI失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeUseAiModel
// @description: 修改使用的AI模型
// @param ctx
func ChangeUseAiModel(ctx *gin.Context) {
	var p changeUseAiModelParam
	if err := ctx.ShouldBind(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`ai_model`", p.Model).Error
	if err != nil {
		log.Printf("修改【%s】的AI模型失败：%s", p.WxId, err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeUseAiAssistant
// @description: 修改使用的AI助手
// @param ctx
func ChangeUseAiAssistant(ctx *gin.Context) {
	// 此处复用一下结构体
	var p changeUseAiModelParam
	if err := ctx.ShouldBind(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`prompt`", p.Model).Error
	if err != nil {
		log.Printf("修改【%s】的AI助手失败：%s", p.WxId, err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeEnableGroupRankStatus
// @description: 修改是否开启水群排行榜
// @param ctx
func ChangeEnableGroupRankStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil || !strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的群Id：%s", p.WxId)

	var record entity.Friend
	if err := client.MySQL.Where("wxid = ?", p.WxId).First(&record).Error; err != nil {
		ctx.String(http.StatusNotFound, "群聊不存在")
		return
	}
	enabled := !record.EnableWeeklyRank
	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Updates(map[string]any{
			"enable_chat_rank":   enabled,
			"enable_weekly_rank": enabled,
		}).Error
	if err != nil {
		log.Printf("修改开启水群排行榜失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeEnableSummaryStatus
// @description: 修改是否开启聊天记录总结
// @param ctx
func ChangeEnableSummaryStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的群Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_summary`", gorm.Expr(" !`enable_summary`")).Error
	if err != nil {
		log.Printf("修改开启聊天记录总结失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeEnableWelcomeStatus
// @description: 修改是否开启迎新
// @param ctx
func ChangeEnableWelcomeStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的群Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_welcome`", gorm.Expr(" !`enable_welcome`")).Error
	if err != nil {
		log.Printf("修改开启迎新失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeEnableCommandStatus
// @description: 修改是否开启指令
// @param ctx
func ChangeEnableCommandStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的群Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_command`", gorm.Expr(" !`enable_command`")).Error
	if err != nil {
		log.Printf("修改指令启用状态失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeSkipGroupRankStatus
// @description: 修改是否跳过水群排行榜
// @param ctx
func ChangeSkipGroupRankStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的群Id：%s -> %s", p.WxId, p.UserId)

	err := client.MySQL.Model(&entity.GroupUser{}).
		Where("group_id = ?", p.WxId).
		Where("wxid = ?", p.UserId).
		Update("`skip_chat_rank`", gorm.Expr(" !`skip_chat_rank`")).Error
	if err != nil {
		log.Printf("修改跳过水群排行榜失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeEnableNewsStatus
// @description: 修改是否开启新闻
// @param ctx
func ChangeEnableNewsStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}
	var record entity.Friend
	if err := client.MySQL.Select("wxid", "enable_news").
		Where("wxid = ?", p.WxId).
		Where("wxid LIKE '%@chatroom'").
		First(&record).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "群聊不存在"})
		return
	}

	enabled := !record.EnableNews
	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("enable_news", enabled).Error
	if err != nil {
		log.Printf("修改早报启用状态失败：%s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

// ChangeEnableHotTopStatus
// @description: 修改是否开启热搜
// @param ctx
func ChangeEnableHotTopStatus(ctx *gin.Context) {
	var p changeStatusParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_hot_top`", gorm.Expr(" !`enable_hot_top`")).Error
	if err != nil {
		log.Printf("修改热榜启用状态失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// AutoClearMembers
// @description: 自动清理群成员
// @param ctx
func AutoClearMembers(ctx *gin.Context) {
	var p autoClearMembers
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`clear_member`", p.Days).Error
	if err != nil {
		log.Printf("修改自动清理群成员阈值失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

// ChangeAiFreeLimit
// @description: 修改AI免费次数
// @param ctx
func ChangeAiFreeLimit(ctx *gin.Context) {
	var p autoClearMembers
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`ai_free_limit`", p.Days).Error
	if err != nil {
		log.Printf("修改AI免费次数失败：%s", err)
		ctx.String(http.StatusInternalServerError, "操作失败: %s", err)
		return
	}

	ctx.String(http.StatusOK, "操作成功")
}

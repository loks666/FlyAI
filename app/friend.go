package app

import (
	"github.com/gin-gonic/gin"
	"go-wechat/client"
	"go-wechat/entity"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// changeStatusParam
// @description: 修改状态用的参数集
type changeStatusParam struct {
	WxId   string `json:"wxId" binding:"required"`
	UserId string `json:"userId"`
}

// changeUseAiModelParam
// @description: 修改使用的AI模型用的参数集
type changeUseAiModelParam struct {
	WxId  string `json:"wxid"`                     // 群Id或微信Id
	Model string `json:"model" binding:"required"` // 模型代码
}

// autoClearMembers
// @description: 自动清理群成员
type autoClearMembers struct {
	WxId string `json:"wxid" binding:"required"` // 群Id
	Days int    `json:"days"`                    // 多少天未发言
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
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的群Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_chat_rank`", gorm.Expr(" !`enable_chat_rank`")).Error
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
		ctx.String(http.StatusBadRequest, "参数错误")
		return
	}
	log.Printf("待修改的Id：%s", p.WxId)

	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", p.WxId).
		Update("`enable_news`", gorm.Expr(" !`enable_news`")).Error
	if err != nil {
		log.Printf("修改早报启用状态失败：%s", err)
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

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/app"
)

// Init
// @description: 初始化路由
// @param g
func Init(g *gin.Engine) {
	g.GET("/", func(ctx *gin.Context) {
		// 重定向到index.html
		ctx.Redirect(302, "/index.html")
	})

	g.GET("/index.html", app.Index)       // 首页
	g.GET("/group.html", app.Group)       // 群组列表
	g.GET("/friend.html", app.Friend)     // 通讯录
	g.GET("/settings.html", app.Settings) // 功能配置

	g.GET("/404.html", app.PageNotFound) // 群组列表

	// 接口
	api := g.Group("/api")
	api.PUT("/monitor/status", app.ChangeEnableMonitorStatus)     // 修改是否监控群消息
	api.GET("/monitor/groups", app.GetEnableMonitorGroups)        // 获取已开启监控的群
	api.PUT("/grouprank/status", app.ChangeEnableGroupRankStatus) // 修改是否开启每周活跃排行榜
	api.PUT("/grouprank/skip", app.ChangeSkipGroupRankStatus)     // 修改成员是否参与排行榜
	api.PUT("/news/status", app.ChangeEnableNewsStatus)           // 修改是否开启每日热点早报
	api.PUT("/dormant/status", app.ChangeEnableDormantStatus)     // 修改是否开启潜水成员统计
	api.GET("/settings", app.GetFeatureSettings)                  // 获取统一功能设置
	api.PUT("/settings/time", app.UpdateFeatureSettingTime)       // 修改统一功能推送时间
	api.GET("/owners", app.GetOwners)                             // 获取主人微信号
	api.POST("/owners", app.CreateOwner)                          // 添加主人微信号
	api.PUT("/owners/:id", app.UpdateOwner)                       // 修改主人微信号
	api.DELETE("/owners/:id", app.DeleteOwner)                    // 删除主人微信号
	api.GET("/group/users", app.GetGroupUsers)                    // 获取群成员列表
	api.POST("/sync/wechat", app.SyncWechatCatalog)               // 手动同步微信群、联系人和当前成员
	api.POST("/tasks/weekly-rank", app.RunWeeklyRank)             // 手动发送指定群的最近7天排行榜
	api.POST("/tasks/weekly-rank/all", app.RunAllWeeklyRanks)     // 一键发送所有已启用群的最近7天排行榜
	api.POST("/tasks/ai-news", app.RunAINews)                     // 手动预览或发送指定群的 AI 早报
	api.POST("/tasks/dormant", app.RunDormantReport)              // 手动预览或发送潜水成员统计
	api.POST("/contacts/message", app.SendDirectMessage)          // 向通讯录联系人私聊发送消息
}

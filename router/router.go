package router

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/app"
	"github.com/loks666/FlyAI/webui"
	"mime"
	"net/http"
	"path"
	"strings"
)

func serveWebPage(assets *webui.Assets, name string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		content, err := assets.Read(name)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "管理页面读取失败")
			return
		}
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", content)
	}
}

// WebFallback serves Next.js static-export data files and extensionless page
// paths that are requested during client-side navigation.
func WebFallback(assets *webui.Assets) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/api") {
			ctx.String(http.StatusNotFound, "接口不存在")
			return
		}

		name := strings.TrimPrefix(path.Clean(ctx.Request.URL.Path), "/")
		candidates := []string{name}
		if name != "" && path.Ext(name) == "" {
			candidates = append(candidates, name+".html", path.Join(name, "index.html"))
		}
		for _, candidate := range candidates {
			content, err := assets.Read(candidate)
			if err != nil {
				continue
			}
			contentType := mime.TypeByExtension(path.Ext(candidate))
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			ctx.Data(http.StatusOK, contentType, content)
			return
		}

		content, err := assets.Read("404.html")
		if err != nil {
			ctx.String(http.StatusNotFound, "页面不存在")
			return
		}
		ctx.Data(http.StatusNotFound, "text/html; charset=utf-8", content)
	}
}

// Init
// @description: 初始化路由
// @param g
func Init(g *gin.Engine, assets *webui.Assets) {
	g.GET("/", serveWebPage(assets, "index.html"))
	g.GET("/index.html", serveWebPage(assets, "index.html"))
	g.GET("/group.html", serveWebPage(assets, "group.html"))
	g.GET("/friend.html", serveWebPage(assets, "friend.html"))
	g.GET("/settings.html", serveWebPage(assets, "settings.html"))
	g.GET("/group", serveWebPage(assets, "group.html"))
	g.GET("/friend", serveWebPage(assets, "friend.html"))
	g.GET("/settings", serveWebPage(assets, "settings.html"))
	g.GET("/404.html", serveWebPage(assets, "404.html"))
	g.GET("/favicon.ico", func(ctx *gin.Context) {
		content, err := assets.Read("flyai-logo.jpg")
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Data(http.StatusOK, "image/jpeg", content)
	})
	g.GET("/flyai-logo.jpg", func(ctx *gin.Context) {
		content, err := assets.Read("flyai-logo.jpg")
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Data(http.StatusOK, "image/jpeg", content)
	})

	// 接口
	api := g.Group("/api")
	api.GET("/dashboard", app.GetWebDashboard)                     // 新版管理端运行概览
	api.GET("/groups", app.GetWebGroups)                           // 新版管理端群组列表
	api.GET("/contacts", app.GetWebContacts)                       // 新版管理端通讯录
	api.PUT("/monitor/status", app.ChangeEnableMonitorStatus)      // 修改是否监控群消息
	api.PUT("/group-feature/status", app.ChangeGroupFeatureStatus) // 修改指定群的自动推送开关
	api.GET("/monitor/groups", app.GetEnableMonitorGroups)         // 获取已开启监控的群
	api.PUT("/grouprank/status", app.ChangeEnableGroupRankStatus)  // 修改是否开启每周活跃排行榜
	api.PUT("/grouprank/skip", app.ChangeSkipGroupRankStatus)      // 修改成员是否参与排行榜
	api.PUT("/news/status", app.ChangeEnableNewsStatus)            // 修改是否开启每日热点早报
	api.PUT("/dormant/status", app.ChangeEnableDormantStatus)      // 修改是否开启潜水成员统计
	api.GET("/settings", app.GetFeatureSettings)                   // 获取统一功能设置
	api.PUT("/settings/time", app.UpdateFeatureSettingTime)        // 修改统一功能推送时间
	api.GET("/owners", app.GetOwners)                              // 获取主人微信号
	api.POST("/owners", app.CreateOwner)                           // 添加主人微信号
	api.PUT("/owners/:id", app.UpdateOwner)                        // 修改主人微信号
	api.DELETE("/owners/:id", app.DeleteOwner)                     // 删除主人微信号
	api.GET("/bridge/status", app.GetBridgeStatus)                 // 手动刷新微信桥接器状态
	api.PUT("/bridge/settings", app.UpdateBridgeSetting)           // 保存机器人微信号
	api.POST("/bridge/start", app.StartBridge)                     // 启动并绑定微信桥接器
	api.POST("/bridge/stop", app.StopBridge)                       // 停止微信桥接器
	api.GET("/group/users", app.GetGroupUsers)                     // 获取群成员列表
	api.POST("/sync/wechat", app.SyncWechatCatalog)                // 手动同步微信群、联系人和当前成员
	api.POST("/tasks/weekly-rank", app.RunWeeklyRank)              // 手动发送指定群的最近7天排行榜
	api.POST("/tasks/weekly-rank/all", app.RunAllWeeklyRanks)      // 一键发送所有已启用群的最近7天排行榜
	api.POST("/tasks/ai-news", app.RunAINews)                      // 手动预览或发送指定群的 AI 早报
	api.POST("/tasks/dormant", app.RunDormantReport)               // 手动预览或发送潜水成员统计
	api.POST("/tasks/group-feature", app.RunGroupFeature)          // 手动向指定群发送某一项推送
	api.POST("/contacts/message", app.SendDirectMessage)           // 向通讯录联系人私聊发送消息
}

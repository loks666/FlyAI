package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-wechat/config"
	"go-wechat/service"
	"net/http"
)

// Index
// @description: 首页
// @param ctx
func Index(ctx *gin.Context) {
	var result = gin.H{
		"msg": "success",
	}
	// 取出所有好友列表
	friends, groups, err := service.GetAllFriend()
	if err != nil {
		result["msg"] = fmt.Sprintf("数据获取失败: %s", err.Error())
	}
	var in, notIn int
	for _, d := range friends {
		if d.IsOk {
			in++
		} else {
			notIn++
		}
	}
	result["friendCount"] = in
	result["friendWithoutCount"] = notIn

	var gin, gnotIn int
	for _, d := range groups {
		if d.IsOk {
			gin++
		} else {
			gnotIn++
		}
	}
	result["groupCount"] = gin
	result["groupWithoutCount"] = gnotIn

	result["vnc"] = config.Conf.Wechat.VncUrl
	result["isVnc"] = config.Conf.Wechat.VncUrl != ""
	result["aiModels"] = config.Conf.Ai.Models

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

	// 取出所有好友列表
	friends, _, err := service.GetAllFriend()
	if err != nil {
		result["msg"] = fmt.Sprintf("数据获取失败: %s", err.Error())
	}
	result["friends"] = friends
	result["vnc"] = config.Conf.Wechat.VncUrl
	result["aiModels"] = config.Conf.Ai.Models
	result["assistant"], _ = service.GetAllAiAssistant()
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
	// 取出所有好友列表
	_, groups, err := service.GetAllFriend()
	if err != nil {
		result["msg"] = fmt.Sprintf("数据获取失败: %s", err.Error())
	}
	result["groups"] = groups
	result["vnc"] = config.Conf.Wechat.VncUrl
	result["aiModels"] = config.Conf.Ai.Models
	result["assistant"], _ = service.GetAllAiAssistant()

	// 渲染页面
	ctx.HTML(http.StatusOK, "group.html", result)
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
	ctx.HTML(http.StatusOK, "404.html", nil)
}

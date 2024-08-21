package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-wechat/model/vo"
	"go-wechat/service"
	"go-wechat/utils"
	"log"
	"net/http"
)

// SendAiSummary
// @description: 发送AI摘要
// @param ctx
func SendAiSummary(ctx *gin.Context) {
	// 获取群Id
	groupId := ctx.Query("id")
	if groupId == "" {
		ctx.String(http.StatusForbidden, "群Id不能为空")
		return
	}

	// 取出群名称
	groupInfo, err := service.GetFriendInfoById(groupId)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "获取群信息失败")
		return
	}

	// 获取对话记录
	var records []vo.TextMessageItem
	if records, err = service.GetTextMessagesById(groupId); err != nil {
		log.Printf("获取群[%s]对话记录失败, 错误信息: %v", groupId, err)
		ctx.String(http.StatusInternalServerError, "获取群对话记录失败")
		return
	}
	if len(records) < 10 {
		ctx.String(http.StatusForbidden, "群对话记录不足10条，建议自己看")
		return
	}

	// 组装对话记录为字符串
	var replyMsg string
	replyMsg, err = utils.GetAiSummary(groupInfo.Nickname, records)
	if err != nil {
		log.Printf("群聊记录总结失败: %v", err.Error())
		ctx.String(http.StatusInternalServerError, "群聊消息总结失败，错误信息: "+err.Error())
		return
	}

	replyMsg = fmt.Sprintf("#昨日消息总结\n又是一天过去了，让我们一起来看看昨儿群友们都聊了什么有趣的话题吧~\n\n%s", replyMsg)
	log.Printf("群[%s]对话记录总结成功，总结内容: %s", groupInfo.Nickname, replyMsg)
	_ = utils.SendMessage(groupId, "", replyMsg, 0)
	ctx.String(http.StatusOK, "操作完成")
}

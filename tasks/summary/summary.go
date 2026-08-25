package summary

import (
	"fmt"
	"github.com/loks666/FlyAI/model/vo"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"time"
)

// AiSummary
// @description: AI总结群聊记录
func AiSummary() {
	groups, err := service.GetAllEnableSummary()
	if err != nil {
		log.Printf("获取启用了聊天排行榜的群组失败, 错误信息: %v", err)
		return
	}

	for _, group := range groups {
		// 记录开始时间
		var start = time.Now()

		// 获取对话记录
		var records []vo.TextMessageItem
		if records, err = service.GetTextMessagesById(group.Wxid); err != nil {
			log.Printf("获取群[%s]对话记录失败, 错误信息: %v", group.Wxid, err)
			continue
		}
		if len(records) < 100 {
			log.Printf("群[%s]对话记录不足100条，跳过总结", group.Wxid)
			continue
		}
		// 组装对话记录为字符串
		var replyMsg string
		replyMsg, err = utils.GetAiSummary(group.Nickname, records)
		if err != nil {
			log.Printf("群聊记录总结失败: %v", err.Error())
			_ = utils.SendMessage(group.Wxid, "", "#昨日消息总结\n\n群聊消息总结失败，错误信息: "+err.Error(), 0)
			continue
		}

		replyMsg = fmt.Sprintf("#昨日消息总结\n又是一天过去了，让我们一起来看看昨儿群友们都聊了什么有趣的话题吧~\n\n%s", replyMsg)
		//log.Printf("群[%s]对话记录总结成功，总结内容: %s", group.Wxid, replyMsg)
		_ = utils.SendMessage(group.Wxid, "", replyMsg, 0)

		// 判断耗时是否达到一分钟，不足就等待
		if used := time.Now().Sub(start); used < time.Minute {
			time.Sleep(time.Minute - used)
		}
	}
}

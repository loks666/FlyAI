package command

import (
	"fmt"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
)

// AiCmd
// @description: AI指令
// @param userId
// @param groupUserId
// @param cmd
func AiCmd(userId, groupUserId, cmd string) {
	// 判断发信人是不是群主
	can := false
	if strings.Contains(userId, "@chatroom") {
		// 判断是不是群主
		err := client.MySQL.Model(&entity.GroupUser{}).
			Where("group_id = ?", userId).
			Where("wxid = ?", groupUserId).
			Pluck("is_admin", &can).Error
		if err != nil {
			log.Printf("查询群主失败: %v", err)
			return
		}
	}
	if !can {
		utils.SendMessage(userId, groupUserId, "您不是群主，无法使用指令", 0)
		return
	}

	var err error
	replyMsg := "操作成功"

	switch cmd {
	case "enable", "启用", "打开":
		err = setAiEnable(userId, true)
	case "disable", "停用", "禁用", "关闭":
		err = setAiEnable(userId, false)
	default:
		replyMsg = "指令错误"
	}
	if err != nil {
		log.Printf("AI指令执行失败: %v", err)
		replyMsg = fmt.Sprintf("指令执行错误: %v", err)
	}
	utils.SendMessage(userId, groupUserId, replyMsg, 0)
}

// setAiEnable
// @description: 设置AI启用状态
// @param userId
// @param enable
// @return err
func setAiEnable(userId string, enable bool) (err error) {
	// 更新
	err = client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", userId).
		Update("enable_ai", enable).Error
	return
}

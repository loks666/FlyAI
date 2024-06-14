package cleargroupuser

import (
	"fmt"
	"go-wechat/client"
	"go-wechat/entity"
	"go-wechat/service"
	"go-wechat/utils"
	"log"
	"strings"
)

// ClearGroupUser
// @description: 清理群成员
func ClearGroupUser() {
	groups, err := service.GetAllEnableClearGroup()
	if err != nil {
		log.Printf("获取启用了聊天排行榜的群组失败, 错误信息: %v", err)
		return
	}

	for _, group := range groups {
		// 获取需要清理的群成员Id
		members := getNeedDeleteMembers(group.Wxid, group.ClearMember)
		memberCount := len(members)
		log.Printf("群[%s(%s)]需要清理的成员数量: %d", group.Nickname, group.Wxid, memberCount)
		if memberCount < 1 {
			continue
		}
		var memberMap = make(map[string]string)
		var deleteIds = make([]string, 0)
		for _, member := range members {
			deleteIds = append(deleteIds, member.Wxid)
			// 昵称为空，取id后4位
			if member.Nickname == "" {
				member.Nickname = "无名氏_" + member.Wxid[len(member.Wxid)-4:]
			}
			memberMap[member.Nickname] = member.LastActive.Format("2006-01-02 15:04:05")
		}
		// 调用接口
		utils.DeleteGroupMember(group.Wxid, strings.Join(deleteIds, ","), 0)
		// 发送通知到群里
		ms := make([]string, 0)
		for k, v := range memberMap {
			ms = append(ms, fmt.Sprintf("昵称：%s\n最后活跃时间：%s", k, v))
		}
		msg := fmt.Sprintf("#清理群成员\n\n很遗憾地通知各位，就在刚刚，有%d名群友因活跃度不够暂时离开了我们，希望还健在的群友引以为戒、保持活跃！\n\n详细信息: \n%s",
			memberCount, strings.Join(ms, "\n"))
		utils.SendMessage(group.Wxid, "", msg, 0)
	}
}

// getNeedDeleteMembers
// @description: 获取需要删除的群成员
// @param groupId 群Id
// @param days 需要清理的未活跃的天数
// @return members
func getNeedDeleteMembers(groupId string, days int) (members []entity.GroupUser) {
	err := client.MySQL.Model(&entity.GroupUser{}).Where("group_id = ?", groupId).
		Where("is_member IS TRUE").
		Where("DATEDIFF( NOW(), last_active ) >= ?", days).
		Order("last_active DESC").
		Find(&members).Error
	if err != nil {
		log.Printf("获取需要清理的群成员失败, 错误信息: %v", err)
	}
	return
}

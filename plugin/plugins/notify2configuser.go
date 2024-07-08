package plugins

import (
	"fmt"
	"go-wechat/config"
	"go-wechat/plugin"
	"go-wechat/utils"
)

// NotifyInvitationJoinGroup
// @description: 通知邀请入群消息到配置用户
// @param m
func NotifyInvitationJoinGroup(m *plugin.MessageContext) {
	// 如果是邀请进群，推送到配置的用户
	if flag, dec := m.IsInvitationJoinGroup(); flag {
		for _, user := range config.Conf.System.NewFriendNotify.ToUser {
			if user != "" {
				// 发送一条新消息
				dec = fmt.Sprintf("#邀请入群提醒\n\n%s", dec)
				utils.SendMessage(user, "", dec, 0)
			}
		}
	}
}

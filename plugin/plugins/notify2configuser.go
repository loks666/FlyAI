package plugins

import (
	"fmt"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/plugin"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
)

// NotifyInvitationJoinGroup
// @description: 通知邀请入群消息到配置用户
// @param m
func NotifyInvitationJoinGroup(m *plugin.MessageContext) {
	// 先回复一条固定消息
	if conf, e := config.Conf.Resource["invitation-join-group"]; e {
		// 发送一条新消息
		switch conf.Type {
		case "text":
			// 文字类型
			_ = utils.SendMessage(m.FromUser, "", conf.Path, 0)
		case "image":
			// 图片类型
			utils.SendImage(m.FromUser, conf.Path, 0)
		case "emotion":
			// 表情类型
			utils.SendEmotion(m.FromUser, conf.Path, 0)
		}
	}

	// 推送到配置的用户
	_, dec := m.IsInvitationJoinGroup()
	for _, user := range config.Conf.System.NewFriendNotify.ToUser {
		if user != "" {
			// 发送一条新消息
			dec = fmt.Sprintf("#邀请入群提醒\n\n%s", dec)
			_ = utils.SendMessage(user, "", dec, 0)
		}
	}
}

// NotifyRemoveFromChatroom
// @description: 被移除群聊通知到配置用户
// @param m
func NotifyRemoveFromChatroom(m *plugin.MessageContext) {
	// 调用一下退出群聊接口，防止被移出后还能从好友列表接口拉到相关信息
	utils.QuitChatroom(m.FromUser, 0)

	// 取出群名称
	groupInfo, err := service.GetFriendInfoById(m.FromUser)
	if err != nil {
		return
	}
	// 组装消息
	msg := fmt.Sprintf("#移除群聊提醒\n\n群Id: %s\n群名称: %s\n事件: %s", m.FromUser, groupInfo.Nickname, m.Content)

	for _, user := range config.Conf.System.NewFriendNotify.ToUser {
		if user != "" {
			// 发送一条新消息
			_ = utils.SendMessage(user, "", msg, 0)
		}
	}

	// 修改is_ok状态
	client.MySQL.Model(&entity.Friend{}).Where("wxid = ?", m.FromUser).Update("is_ok", false)
}

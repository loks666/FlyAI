package plugins

import (
	"fmt"
	"go-wechat/client"
	"go-wechat/config"
	"go-wechat/model/entity"
	"go-wechat/plugin"
	"go-wechat/service"
	"go-wechat/utils"
)

// NotifyInvitationJoinGroup
// @description: 通知邀请入群消息到配置用户
// @param m
func NotifyInvitationJoinGroup(m *plugin.MessageContext) {
	// 先回复一条固定句子
	_ = utils.SendMessage(m.FromUser, m.GroupUser, "您的邀请消息已收到啦，正在通知我的主人来同意请求。在我加群之后将会进行初始化操作，直到收到我主动发送的消息就是初始化完成咯，在那之前请耐心等待喔~", 0)

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

package initialization

import (
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/model/dto"
	plugin "github.com/loks666/FlyAI/plugin"
	"github.com/loks666/FlyAI/plugin/plugins"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/types"
)

// Plugin
// @description: 初始化插件
func Plugin() {
	// 定义一个处理器
	dispatcher := plugin.NewMessageMatchDispatcher()
	// 设置为异步处理
	dispatcher.SetAsync(true)

	// 注册插件

	// 保存消息进数据库
	dispatcher.RegisterHandler(func(*dto.Message) bool {
		return true
	}, plugins.SaveToDb)

	// 通知邀请入群消息到配置用户
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		flag, _ := m.IsInvitationJoinGroup()
		return flag && !m.IsGroup()
	}, plugins.NotifyInvitationJoinGroup)
	// 被移除群聊通知到配置用户
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return m.IsRemoveFromChatroom()
	}, plugins.NotifyRemoveFromChatroom)
	// 响应好友添加成功消息
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return m.IsNewFriendAdd() || m.IsJoinToGroup() || m.IsOldFriendBack()
	}, plugins.ReplyNewFriend)

	// 私聊指令消息
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return m.IsGroup() && m.Type == types.MsgTypeText && service.IsOwnerWxid(m.GroupUser) && ownerCommandMatches(m)
	}, plugins.OwnerCommand)

	// 普通指令只允许私聊；群聊仅开放绑定主人的三个固定命令。
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return m.IsPrivateText() && m.CleanContentStartWith("/") && service.CheckIsEnableCommand(m.FromUser)
	}, plugins.Command)

	// AI消息插件
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		// AI 尚未接入群主人命令；普通群消息不触发 AI。
		return m.IsPrivateText()
	}, plugins.AI)

	// 欢迎新成员
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return m.IsNewUserJoin()
	}, plugins.WelcomeNew)

	// 注册消息处理器
	handler := plugin.DispatchMessage(dispatcher)
	current.SetRobotMessageHandler(func(message *dto.Message) {
		if message.IsBackfill {
			plugins.SaveDTOToDb(message)
			return
		}
		if message.IsGroup() && !service.IsGroupMonitorEnabled(message.FromUser) {
			return
		}
		handler(message)
	})
}

func ownerCommandMatches(m *dto.Message) bool {
	_, ok := service.ParseOwnerCommand(m.Content, current.GetRobotInfo().Name)
	return ok
}

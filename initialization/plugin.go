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

	// 群聊中的主人指令消息
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return isOwnerCommandMessage(m)
	}, plugins.OwnerCommand)

	// 非管理员在群里 @机器人时，执行按发送者 wxid 隔离的回复状态机。
	dispatcher.RegisterHandler(func(m *dto.Message) bool {
		return isNonOwnerMentionMessage(m)
	}, plugins.AttackReply)

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
		// Complete the persistence attempt before acknowledging the bridge, then
		// queue live business handlers. Historical rows never trigger replies.
		plugins.SaveDTOToDb(message)
		if message.IsBackfill {
			return
		}
		// 第二次非管理员回复成功后，静默状态按发送者 wxid 跨群、
		// 跨私聊生效；消息仍已在上方完成入库。
		if message.IsPrivateText() && service.ShouldIgnoreNonOwnerMessage(message.FromUser) {
			return
		}
		// 主人手动指令与群的定时监控开关彻底分离：即使群未开启
		// 监控，绑定主人仍可以在群里 @机器人执行固定命令。
		if message.IsGroup() && !service.IsGroupMonitorEnabled(message.FromUser) {
			if !isOwnerCommandMessage(message) && !isNonOwnerMentionMessage(message) {
				// Keep the history complete for on-demand owner reports, but do not
				// run any automatic response plugin for a disabled group.
				return
			}
		}
		handler(message)
	})
}

func isOwnerCommandMessage(m *dto.Message) bool {
	return m.IsGroup() && m.Type == types.MsgTypeText && service.IsOwnerWxid(m.GroupUser) && ownerCommandMatches(m)
}

func isNonOwnerMentionMessage(m *dto.Message) bool {
	robot := current.GetRobotInfo()
	return m.IsGroup() &&
		m.Type == types.MsgTypeText &&
		m.GroupUser != robot.WxId &&
		!service.IsOwnerWxid(m.GroupUser) &&
		service.IsRobotMention(m.Content, robot.Name)
}

func ownerCommandMatches(m *dto.Message) bool {
	robotName := current.GetRobotInfo().Name
	if service.HasRobotCommand(m.Content, robotName) {
		return true
	}
	_, _, ok := service.ParseOwnerAttackTargets(m.Content, robotName)
	return ok
}

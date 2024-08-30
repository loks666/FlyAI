package plugins

import (
	"go-wechat/plugin"
	"go-wechat/service"
	"go-wechat/utils"
	"time"
)

// ReplyNewFriend
// @description: 响应好友添加成功消息
// @param m
func ReplyNewFriend(m *plugin.MessageContext) {
	if m.IsNewFriendAdd() || m.IsJoinToGroup() {
		_ = utils.SendMessage(m.FromUser, m.GroupUser, "AI正在初始化，请稍等几分钟，初始化完成之后我将主动告知您。", 0)
	}
	if m.IsOldFriendBack() {
		_ = utils.SendMessage(m.FromUser, "", "嘿，我的朋友，你为何要离我而去？又为何去而复返？", 0)
	}

	go func() {
		// 等待5秒
		time.Sleep(5 * time.Second)
		// 刷新好友列表
		service.SyncFriend()
		// 如果是加入群，刷新群成员列表
		if m.IsJoinToGroup() {
			service.SyncGroupMembers(m.FromUser)
		}
	}()
}

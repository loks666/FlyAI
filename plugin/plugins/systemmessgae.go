package plugins

import (
	"go-wechat/plugin"
	"go-wechat/utils"
	"strings"
)

// ReplyNewFriend
// @description: 响应好友添加成功消息
// @param m
func ReplyNewFriend(m *plugin.MessageContext) {
	isNewFriend := strings.HasPrefix(m.Content, "你已添加了") && strings.HasSuffix(m.Content, "，现在可以开始聊天了。")
	isNewChatroom := strings.Contains(m.Content, "\"邀请你加入了群聊，群聊参与人还有：")
	if isNewFriend || isNewChatroom {
		utils.SendMessage(m.FromUser, m.GroupUser, "AI正在初始化，请稍等几分钟，初始化完成之后我将主动告知您。", 0)
	}
}

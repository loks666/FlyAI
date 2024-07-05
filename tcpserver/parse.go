package tcpserver

import (
	"encoding/json"
	"go-wechat/common/current"
	"go-wechat/config"
	"go-wechat/model"
	"go-wechat/types"
	"go-wechat/utils"
	"log"
	"net"
	"strings"
)

// parse
// @description: 解析消息
// @param msg
func parse(remoteAddr net.Addr, msg []byte) {
	var m model.Message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Printf("[%s]消息解析失败： %v", remoteAddr, err)
		log.Printf("[%s]消息内容： %d -> %v", remoteAddr, len(msg), string(msg))
		return
	}
	// 记录原始数据
	m.Raw = string(msg)

	// 提取出群成员信息
	// Sys类型的消息正文不包含微信 Id，所以不需要处理
	if m.IsGroup() && m.Type != types.MsgTypeSys {
		// 群消息，处理一下消息和发信人
		groupUser := strings.Split(m.Content, "\n")[0]
		groupUser = strings.ReplaceAll(groupUser, ":", "")
		// 如果两个id一致，说明是系统发的
		if m.FromUser != groupUser {
			m.GroupUser = groupUser
		}
		// 用户的操作单独提出来处理一下
		m.Content = strings.Join(strings.Split(m.Content, "\n")[1:], "\n")
	}
	log.Printf("%s\n消息来源: %s\n群成员: %s\n消息类型: %v\n消息内容: %s", remoteAddr, m.FromUser, m.GroupUser, m.Type, m.Content)

	// 如果是邀请进群，推送到配置的用户
	if flag, dec := m.IsInvitationJoinGroup(); flag {
		for _, user := range config.Conf.System.NewFriendNotify.ToUser {
			if user != "" {
				// 发送一条新消息
				utils.SendMessage(user, "", dec, 0)
			}
		}
	}

	// 插件不为空，开始执行
	if p := current.GetRobotMessageHandler(); p != nil {
		p(&m)
	}

}

package plugins

import (
	"go-wechat/model/entity"
	"go-wechat/plugin"
	"go-wechat/service"
	"time"
)

// SaveToDb
// @description: 保存消息到数据库
// @param m
func SaveToDb(m *plugin.MessageContext) {
	var ent entity.Message
	ent.MsgId = m.MsgId
	ent.CreateTime = m.CreateTime
	ent.CreateAt = time.Unix(int64(m.CreateTime), 0)
	ent.Content = m.Content
	ent.FromUser = m.FromUser
	ent.GroupUser = m.GroupUser
	ent.ToUser = m.ToUser
	ent.Type = m.Type
	ent.DisplayFullContent = m.DisplayFullContent
	ent.Raw = m.Raw
	// 保存入库
	service.SaveMessage(ent)
}

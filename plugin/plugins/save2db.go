package plugins

import (
	"github.com/loks666/FlyAI/model/dto"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/plugin"
	"github.com/loks666/FlyAI/service"
	"time"
)

// SaveToDb
// @description: 保存消息到数据库
// @param m
func SaveToDb(m *plugin.MessageContext) {
	SaveDTOToDb(m.Message)
}

// SaveDTOToDb persists a parsed live or historical bridge message without
// invoking any response-oriented plugin.
func SaveDTOToDb(m *dto.Message) {
	ent := messageEntity(m)
	// 保存入库
	service.SaveMessage(ent)
}

func messageEntity(m *dto.Message) entity.Message {
	var ent entity.Message
	ent.MsgId = m.MsgId
	ent.CreateTime = m.CreateTime
	ent.CreateAt = time.Unix(int64(m.CreateTime), 0)
	ent.Content = m.Content
	ent.FromUser = m.FromUser
	ent.GroupUser = m.GroupUser
	ent.SenderNickname = m.SenderNickname
	ent.SenderGroupNickname = m.SenderGroupNickname
	ent.SenderWechatAccount = m.SenderWechatAccount
	ent.GroupNickname = m.GroupNickname
	ent.ToUser = m.ToUser
	ent.Type = m.Type
	ent.DisplayFullContent = m.DisplayFullContent
	ent.Raw = m.Raw
	return ent
}

// SaveDTOBatchToDb persists historical messages with one batched upsert.
func SaveDTOBatchToDb(messages []*dto.Message) error {
	entities := make([]entity.Message, 0, len(messages))
	for _, message := range messages {
		entities = append(entities, messageEntity(message))
	}
	return service.SaveMessagesBatch(entities)
}

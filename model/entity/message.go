package entity

import (
	"go-wechat/types"
	"time"
)

// Message
// @description: 消息数据库结构体
type Message struct {
	MsgId              int64             `gorm:"primaryKey"` // 消息Id
	CreateTime         int               // 发送时间戳
	CreateAt           time.Time         // 发送时间
	Type               types.MessageType // 消息类型
	Content            string            // 内容
	DisplayFullContent string            // 显示的完整内容
	FromUser           string            // 发送者
	GroupUser          string            // 群成员
	ToUser             string            // 接收者
	Raw                string            // 原始通知字符串
}

func (Message) TableName() string {
	return "t_message"
}

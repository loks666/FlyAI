package service

import (
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/model/vo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"strconv"
	"time"
)

// SaveMessage
// @description: 消息入库
// @param msg
func SaveMessage(msg entity.Message) {
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SAVE")); flag {
		return
	}

	// 检查消息是否存在。桥接器重复回填时允许修正发送者映射，
	// 这样微信内部数字 contact.id 被解析成 wxid 后无需清空历史数据。
	var old entity.Message
	err := client.MySQL.Where("msg_id = ?", msg.MsgId).First(&old).Error
	if err == nil {
		if old.GroupUser != msg.GroupUser || old.Type != msg.Type || old.Content != msg.Content ||
			old.SenderNickname != msg.SenderNickname ||
			old.SenderGroupNickname != msg.SenderGroupNickname ||
			old.SenderWechatAccount != msg.SenderWechatAccount ||
			old.GroupNickname != msg.GroupNickname {
			err = client.MySQL.Model(&entity.Message{}).
				Where("msg_id = ?", msg.MsgId).
				Updates(map[string]any{
					"create_time":           msg.CreateTime,
					"create_at":             msg.CreateAt,
					"type":                  msg.Type,
					"content":               msg.Content,
					"display_full_content":  msg.DisplayFullContent,
					"from_user":             msg.FromUser,
					"group_user":            msg.GroupUser,
					"sender_nickname":       msg.SenderNickname,
					"sender_group_nickname": msg.SenderGroupNickname,
					"sender_wechat_account": msg.SenderWechatAccount,
					"group_nickname":        msg.GroupNickname,
					"to_user":               msg.ToUser,
					"raw":                   msg.Raw,
				}).Error
			if err != nil {
				log.Printf("修正消息发送者失败, 消息Id: %d, 错误信息: %v", msg.MsgId, err)
			}
		}
		return
	}
	if err != gorm.ErrRecordNotFound {
		log.Printf("检查消息是否存在失败, 错误信息: %v", err)
		return
	}
	err = client.MySQL.Create(&msg).Error
	if err != nil {
		log.Printf("消息入库失败, 错误信息: %v", err)
		return
	}

	// 更新最后活跃时间
	// 只处理收到的消息
	if msg.MsgId > 1 {
		go updateLastActive(msg)
	}
}

// SaveMessagesBatch upserts a historical replay in one database transaction.
// It deliberately avoids the per-message existence query used by live traffic.
func SaveMessagesBatch(messages []entity.Message) error {
	if len(messages) == 0 {
		return nil
	}
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SAVE")); flag {
		return nil
	}

	columns := []string{
		"create_time", "create_at", "type", "content", "display_full_content",
		"from_user", "group_user", "sender_nickname", "sender_group_nickname",
		"sender_wechat_account", "group_nickname", "to_user", "raw",
	}
	return client.MySQL.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "msg_id"}},
			DoUpdates: clause.AssignmentColumns(columns),
		}).CreateInBatches(messages, len(messages)).Error; err != nil {
			return err
		}
		return updateBatchLastActive(tx, messages)
	})
}

type groupMemberKey struct {
	groupID string
	wxid    string
}

func updateBatchLastActive(tx *gorm.DB, messages []entity.Message) error {
	groupLatest := make(map[string]time.Time)
	memberLatest := make(map[groupMemberKey]time.Time)
	for _, msg := range messages {
		if msg.MsgId <= 1 {
			continue
		}
		if msg.CreateAt.After(groupLatest[msg.FromUser]) {
			groupLatest[msg.FromUser] = msg.CreateAt
		}
		if msg.GroupUser != "" {
			key := groupMemberKey{groupID: msg.FromUser, wxid: msg.GroupUser}
			if msg.CreateAt.After(memberLatest[key]) {
				memberLatest[key] = msg.CreateAt
			}
		}
	}
	for groupID, activeAt := range groupLatest {
		if err := tx.Model(&entity.Friend{}).
			Where("wxid = ?", groupID).
			Update("last_active", gorm.Expr("GREATEST(last_active, ?)", activeAt)).Error; err != nil {
			return err
		}
	}
	for key, activeAt := range memberLatest {
		if err := tx.Model(&entity.GroupUser{}).
			Where("group_id = ? AND wxid = ?", key.groupID, key.wxid).
			Update("last_active", gorm.Expr("GREATEST(last_active, ?)", activeAt)).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetTextMessagesById
// @description: 根据群id或者用户Id获取消息
// @param id
// @return records
// @return err
func GetTextMessagesById(id string) (records []vo.TextMessageItem, err error) {
	// APP消息类型
	appMsgList := []string{"57", "4", "5", "6"}
	// 这个查询子句抽出来写，方便后续扩展
	selectStr := `CASE
		WHEN tm.type = 49 THEN
	CASE
			WHEN EXTRACTVALUE ( tm.content, "/msg/appmsg/type" ) = '57' THEN
			EXTRACTVALUE ( tm.content, "/msg/appmsg/title" )
			WHEN EXTRACTVALUE ( tm.content, "/msg/appmsg/type" ) = '5' THEN
			CONCAT("网页分享消息，标题: ", EXTRACTVALUE (tm.content, "/msg/appmsg/title"), "，描述：", EXTRACTVALUE (tm.content, "/msg/appmsg/des"))
			WHEN EXTRACTVALUE ( tm.content, "/msg/appmsg/type" ) = '4' THEN
			CONCAT("网页分享消息，标题: ", EXTRACTVALUE (tm.content, "/msg/appmsg/title"), "，描述：", EXTRACTVALUE (tm.content, "/msg/appmsg/des"))
			WHEN EXTRACTVALUE ( tm.content, "/msg/appmsg/type" ) = '6' THEN
			CONCAT("文件消息，文件名: ", EXTRACTVALUE (tm.content, "/msg/appmsg/title"))

			ELSE EXTRACTVALUE ( tm.content, "/msg/appmsg/des" )
		END ELSE tm.content
	END`

	tx := client.MySQL.
		Table("`t_message` AS tm").
		Joins("LEFT JOIN t_group_user AS tgu ON tm.group_user = tgu.wxid AND tgu.group_id = tm.from_user").
		Select("tgu.nickname", selectStr+" AS message").
		Where("tm.`from_user` = ?", id).
		Where(`(tm.type = 1 OR ( tm.type = 49 AND EXTRACTVALUE ( tm.content, "/msg/appmsg/type" ) IN (?) ))`, appMsgList).
		Where("DATE ( tm.create_at ) = DATE ( CURDATE() - INTERVAL 1 DAY )").
		Where("tm.content NOT LIKE '#昨日水群排行榜%'").
		Where("tm.content NOT LIKE '#昨日消息总结%'").
		Order("tm.create_at ASC")

	err = tx.Find(&records).Error
	return
}

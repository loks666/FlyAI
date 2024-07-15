package service

import (
	"go-wechat/client"
	"go-wechat/model/entity"
	"go-wechat/model/vo"
	"gorm.io/gorm"
	"log"
	"strings"
)

// GetAllFriend
// @description: 取出所有好友
// @return friends
// @return groups
// @return err
func GetAllFriend() (friends, groups []vo.FriendItem, err error) {
	var records []vo.FriendItem
	err = client.MySQL.
		Table("t_friend AS tf").
		//Joins("LEFT JOIN t_message AS tm ON tf.wxid = tm.from_user").
		//Select("tf.*", "MAX(tm.create_at) AS last_active").
		Select("tf.*").
		//Group("tf.wxid").
		Order("tf.last_active DESC").
		Find(&records).Error
	if err != nil {
		return
	}
	for _, record := range records {
		if strings.HasSuffix(record.Wxid, "@chatroom") {
			groups = append(groups, record)
		} else {
			friends = append(friends, record)
		}
	}
	return
}

// GetFriendInfoById
// @description: 通过wxId获取好友信息
// @param wxId
// @return ent
// @return err
func GetFriendInfoById(wxId string) (ent entity.Friend, err error) {
	err = client.MySQL.Where("wxid = ?", wxId).First(&ent).Error
	return
}

// GetAllEnableAI
// @description: 取出所有启用了AI的好友或群组
// @return []entity.Friend
func GetAllEnableAI() (records []entity.Friend, err error) {
	err = client.MySQL.Where("enable_ai = ?", 1).Where("is_ok IS TRUE").Find(&records).Error
	return
}

// GetAllEnableChatRank
// @description: 取出所有启用了聊天排行榜的群组
// @return records
// @return err
func GetAllEnableChatRank() (records []entity.Friend, err error) {
	err = client.MySQL.Where("enable_chat_rank = ?", 1).
		Where("is_ok IS TRUE").
		Where("wxid LIKE '%@chatroom'").
		Find(&records).Error
	return
}

// GetAllEnableSummary
// @description: 取出所有启用了总结的群组
// @return records
// @return err
func GetAllEnableSummary() (records []entity.Friend, err error) {
	err = client.MySQL.Where("enable_summary = ?", 1).
		Where("is_ok IS TRUE").
		Where("wxid LIKE '%@chatroom'").
		Find(&records).Error
	return
}

// GetAllEnableNews
// @description: 取出所有启用了新闻的好友或群组
// @return records
// @return err
func GetAllEnableNews() (records []entity.Friend, err error) {
	err = client.MySQL.Where("enable_news = ?", 1).Where("is_ok IS TRUE").Find(&records).Error
	return
}

// GetAllEnableHotTop
// @description: 获取所有启用了热搜排行榜的好友或群组
// @return records
// @return err
func GetAllEnableHotTop() (records []entity.Friend, err error) {
	err = client.MySQL.Where("enable_hot_top = ?", 1).Where("is_ok IS TRUE").Find(&records).Error
	return
}

// GetAllEnableClearGroup
// @description: 获取所有需要清理成员的群组
// @return records
// @return err
func GetAllEnableClearGroup() (records []entity.Friend, err error) {
	err = client.MySQL.Where("clear_member > 0").Where("is_ok IS TRUE").Find(&records).Error
	return
}

// CheckIsEnableCommand
// @description: 检查用户是否启用了指令
// @param userId
// @return flag
func CheckIsEnableCommand(userId string) (flag bool) {
	var coo int64
	client.MySQL.Model(&entity.Friend{}).Where("enable_command = 1").Where("wxid = ?", userId).Count(&coo)
	return coo > 0
}

// updateLastActive
// @description: 更新最后活跃时间
// @param msg
func updateLastActive(msg entity.Message) {
	var err error
	// 如果是群，更新群成员最后活跃时间
	if strings.HasSuffix(msg.FromUser, "@chatroom") {
		err = client.MySQL.Model(&entity.GroupUser{}).
			Where("group_id = ?", msg.FromUser).
			Where("wxid = ?", msg.GroupUser).
			Update("last_active", msg.CreateAt).Error
		if err != nil {
			log.Printf("更新群成员最后活跃时间失败, 错误信息: %v", err)
		}
	}
	// 更新群或者好友活跃时间
	err = client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", msg.FromUser).
		Update("last_active", msg.CreateAt).Error
	if err != nil {
		log.Printf("更新群或者好友活跃时间失败, 错误信息: %v", err)
	}
}

// UpdateUsedAiTokens
// @description: 更新已使用的AI次数
// @param wxId 微信好友或者群聊Id
// @param tokens 新增的tokens额度
func UpdateUsedAiTokens(wxId string, tokens int) {
	err := client.MySQL.Model(&entity.Friend{}).
		Where("wxid = ?", wxId).
		Update("`used_tokens`", gorm.Expr(" `used_tokens` + ?", tokens)).Error
	if err != nil {
		log.Printf("更新AI使用次数失败, 错误信息: %v", err)
	}
}

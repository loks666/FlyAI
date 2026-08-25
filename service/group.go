package service

import (
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/model/vo"
	"sort"
)

type DormantMember struct {
	Wxid     string `json:"wxid"`
	Nickname string `json:"nickname"`
	Count    int64  `json:"count"`
}

// GetGroupUsersByGroupId
// @description: 根据群Id取出群成员列表
// @param groupId
// @return records
// @return err
func GetGroupUsersByGroupId(groupId string) (records []vo.GroupUserItem, err error) {
	err = client.MySQL.
		Table("t_group_user AS tgu").
		Select("tgu.*").
		Where("tgu.group_id = ?", groupId).
		Order("tgu.join_time DESC").
		Find(&records).Error
	return
}

// CountCurrentGroupMembers returns the distinct current member count.
func CountCurrentGroupMembers(groupId string) (count int64, err error) {
	err = client.MySQL.Model(&entity.GroupUser{}).
		Where("group_id = ?", groupId).
		Where("is_member IS TRUE").
		Distinct("wxid").
		Count(&count).Error
	return
}

// GetDormantGroupMembers returns current members whose user-message total is
// strictly less than threshold. A missing count means zero messages. Group
// nicknames take precedence over personal nicknames.
func GetDormantGroupMembers(groupId string, messageCounts map[string]int64, selfId string, threshold int) (records []DormantMember, err error) {
	tx := client.MySQL.
		Table("t_group_user AS tgu").
		Joins("LEFT JOIN t_friend AS tf ON tf.wxid = tgu.wxid").
		Select(`tgu.wxid,
			COALESCE(NULLIF(MAX(tgu.nickname), ''), NULLIF(MAX(tf.nickname), ''), tgu.wxid) AS nickname`).
		Where("tgu.group_id = ?", groupId)
	if selfId != "" {
		tx = tx.Where("tgu.wxid <> ?", selfId)
	}
	var currentMembers []DormantMember
	err = tx.
		Group("tgu.wxid").
		Having("MAX(tgu.is_member) = 1").
		Order("nickname ASC").
		Find(&currentMembers).Error
	if err != nil {
		return nil, err
	}
	return filterDormantMembers(currentMembers, messageCounts, threshold), nil
}

func filterDormantMembers(currentMembers []DormantMember, messageCounts map[string]int64, threshold int) (records []DormantMember) {
	for _, member := range currentMembers {
		member.Count = messageCounts[member.Wxid]
		if member.Count < int64(threshold) {
			records = append(records, member)
		}
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Count == records[j].Count {
			return records[i].Nickname < records[j].Nickname
		}
		return records[i].Count < records[j].Count
	})
	return
}

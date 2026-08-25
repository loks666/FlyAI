package main

import (
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/initialization"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
	"time"
)

// syncbridge adds group members exposed by the local WeChat 4.x bridge using
// short independent writes. It intentionally avoids the legacy full-sync
// transaction, which becomes lock-prone once thousands of member rows exist.
func main() {
	initialization.InitConfig()

	groups, err := utils.GetFriendList()
	if err != nil {
		log.Fatalf("获取桥接群列表失败: %v", err)
	}

	var added, updated int
	for _, group := range groups {
		if !strings.HasSuffix(group.Wxid, "@chatroom") {
			continue
		}

		ids, memberNicknames, admin, err := utils.GetGroupMembers(group.Wxid)
		if err != nil {
			log.Printf("获取群[%s]成员失败: %v", group.Wxid, err)
			continue
		}
		if len(ids) > 0 {
			now := time.Now().Local()
			if err = client.MySQL.Model(&entity.GroupUser{}).
				Where("group_id = ?", group.Wxid).
				Where("is_member IS TRUE").
				Where("wxid NOT IN (?)", ids).
				Updates(map[string]any{"is_member": false, "leave_time": now}).Error; err != nil {
				log.Printf("标记群[%s]已退出成员失败: %v", group.Wxid, err)
				continue
			}
		}

		var existing []entity.GroupUser
		if err = client.MySQL.Where("group_id = ?", group.Wxid).Find(&existing).Error; err != nil {
			log.Printf("读取群[%s]已有成员失败: %v", group.Wxid, err)
			continue
		}
		known := make(map[string]entity.GroupUser, len(existing))
		for _, member := range existing {
			known[member.Wxid] = member
		}

		for _, wxid := range ids {
			if wxid == "" {
				continue
			}
			preferredNickname := memberNicknames[wxid]
			if oldMember, ok := known[wxid]; ok {
				updates := map[string]any{
					"is_member": true,
					"is_admin":  wxid == admin,
				}
				if preferredNickname != "" && preferredNickname != oldMember.Nickname {
					updates["nickname"] = preferredNickname
				}
				if err = client.MySQL.Model(&entity.GroupUser{}).
					Where("group_id = ?", group.Wxid).
					Where("wxid = ?", wxid).
					Updates(updates).Error; err != nil {
					log.Printf("更新群[%s]成员[%s]失败: %v", group.Wxid, wxid, err)
					continue
				}
				if preferredNickname != "" && preferredNickname != oldMember.Nickname {
					updated++
				}
				continue
			}
			profile, err := utils.GetContactProfile(wxid)
			if err != nil || profile.Wxid == "" {
				continue
			}
			now := time.Now().Local()
			nickname := preferredNickname
			if nickname == "" {
				nickname = profile.Nickname
			}
			member := entity.GroupUser{
				GroupId:    group.Wxid,
				Account:    profile.Account,
				HeadImage:  profile.HeadImage,
				Nickname:   nickname,
				Wxid:       profile.Wxid,
				IsMember:   true,
				IsAdmin:    profile.Wxid == admin,
				JoinTime:   now,
				LastActive: now,
			}
			if err = client.MySQL.Create(&member).Error; err != nil {
				log.Printf("新增群[%s]成员[%s]失败: %v", group.Wxid, wxid, err)
				continue
			}
			known[wxid] = member
			added++
		}
	}

	log.Printf("桥接成员同步完成，新增 %d 条，更新群内昵称 %d 条", added, updated)
}

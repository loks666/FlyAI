package service

import (
	"go-wechat/client"
	"go-wechat/common/types"
	"go-wechat/config"
	"go-wechat/model/entity"
	"go-wechat/utils"
	"log"
	"strings"
	"time"
)

// SyncFriend
// @description: 同步好友列表
func SyncFriend() {
	// 获取好友列表
	friends, err := utils.GetFriendList()
	if err != nil {
		log.Printf("获取好友列表失败: %s", err.Error())
		return
	}
	// 当前获取到的成员Id，用于后续设置is_ok状态
	nowIds := make([]string, 0)
	// 取出已存在的成员信息
	var oldData []entity.Friend
	err = client.MySQL.Find(&oldData).Error
	if err != nil {
		log.Printf("查询好友列表失败: %s", err.Error())
		return
	}
	// 将历史数据整理成map
	oldMap := make(map[string]bool)
	for _, item := range oldData {
		oldMap[item.Wxid] = item.IsOk
	}

	// 新增的成员，用于通知给指定的人
	var notifyMap = make(map[string]string)

	// 开启事务
	tx := client.MySQL.Begin()
	defer tx.Commit()

	// 循环获取到的好友列表
	for _, item := range friends {
		// 填充当前存在的账号
		nowIds = append(nowIds, item.Wxid)

		// 判断是否已经存在
		if _, ok := oldMap[item.Wxid]; ok {
			// 已存在，修改
			pm := map[string]any{
				"nickname":       item.Nickname,
				"custom_account": item.CustomAccount,
				"pinyin":         item.Pinyin,
				"pinyin_all":     item.PinyinAll,
				"is_ok":          true,
			}
			err = tx.Model(&entity.Friend{}).Where("wxid = ?", item.Wxid).Updates(pm).Error
			if err != nil {
				log.Printf("修改好友失败: %s", err.Error())
				continue
			}

			// 如果已存在但是是已退出的群，也通知一下
			if !oldMap[item.Wxid] {
				notifyMap[item.Wxid] = item.Nickname + " #秽土转生"

				// 通知一下，初始化完成
				if conf, e := config.Conf.Resource["introduce"]; e {
					// 发送一条新消息
					switch conf.Type {
					case "text":
						// 文字类型
						_ = utils.SendMessage(item.Wxid, "", conf.Path, 0)
					case "image":
						// 图片类型
						utils.SendImage(item.Wxid, conf.Path, 0)
					case "emotion":
						// 表情类型
						utils.SendEmotion(item.Wxid, conf.Path, 0)
					}
				}
				// 发送配置网页
				if config.Conf.System.Domain != "" {
					title := "欢迎使用微信机器人(切勿转发)"
					desc := "点我可以配置功能喔，提示非微信官方网页，点击继续访问即可"
					url := utils.GetManagerUrl(item.Wxid)
					utils.SendPublicMsg(item.Wxid, title, desc, url, 0)
				}
			}
		} else {
			// 新增
			err = tx.Create(&entity.Friend{
				CustomAccount:  item.CustomAccount,
				Nickname:       item.Nickname,
				Pinyin:         item.Pinyin,
				PinyinAll:      item.PinyinAll,
				Wxid:           item.Wxid,
				IsOk:           true,
				EnableAi:       config.Conf.System.DefaultRule.Ai,
				EnableChatRank: config.Conf.System.DefaultRule.ChatRank,
				EnableSummary:  config.Conf.System.DefaultRule.Summary,
				EnableWelcome:  config.Conf.System.DefaultRule.Welcome,
				EnableNews:     config.Conf.System.DefaultRule.News,
				EnableHotTop:   config.Conf.System.DefaultRule.HotTop,
				AiFreeLimit:    config.Conf.System.DefaultRule.AiFreeLimit,
				ClearMember:    0,
				LastActive:     types.DateTime(time.Now().Local()),
			}).Error
			if err != nil {
				log.Printf("新增好友失败: %s", err.Error())
				continue
			}
			notifyMap[item.Wxid] = item.Nickname
			if conf, e := config.Conf.Resource["introduce"]; e {
				// 发送一条新消息
				switch conf.Type {
				case "text":
					// 文字类型
					_ = utils.SendMessage(item.Wxid, "", conf.Path, 0)
				case "image":
					// 图片类型
					utils.SendImage(item.Wxid, conf.Path, 0)
				case "emotion":
					// 表情类型
					utils.SendEmotion(item.Wxid, conf.Path, 0)
				}
			}
			// 发送配置网页
			if config.Conf.System.Domain != "" {
				title := "欢迎使用微信机器人(切勿转发)"
				desc := "点我可以配置功能喔，提示非微信官方网页，点击继续访问即可"
				url := utils.GetManagerUrl(item.Wxid)
				utils.SendPublicMsg(item.Wxid, title, desc, url, 0)
			}
		}
	}

	// 通知有新成员
	if len(notifyMap) > 0 && config.Conf.System.NewFriendNotify.Enable {
		// 组装成一句话
		msg := []string{"#新好友通知\n"}
		for wxId, nickname := range notifyMap {
			msg = append(msg, "微信Id: "+wxId+"\n昵称: "+nickname)
		}
		for _, user := range config.Conf.System.NewFriendNotify.ToUser {
			if user != "" {
				// 发送一条新消息
				_ = utils.SendMessage(user, "", strings.Join(msg, "\n-------\n"), 0)
			}
		}
	}

	// 清理不在列表中的好友
	clearPm := map[string]any{
		"is_ok": false,
	}
	err = tx.Model(&entity.Friend{}).Where("wxid NOT IN (?)", nowIds).Updates(clearPm).Error
	if err != nil {
		log.Printf("清理好友失败: %s", err.Error())
	}

	log.Println("同步好友列表完成")
}

// SyncGroupMembers
// @description: 同步群成员
// @param wxId
func SyncGroupMembers(wxId string) {
	membersIds, admin, err := utils.GetGroupMembers(wxId)
	if err != nil {
		return
	}

	// 修改不在数组的群成员状态为不在
	pm := map[string]any{
		"is_member":  false,
		"leave_time": time.Now().Local(),
	}
	err = client.MySQL.Model(&entity.GroupUser{}).
		Where("group_id = ?", wxId).
		Where("is_member IS TRUE").
		Where("wxid NOT IN (?)", membersIds).
		Updates(pm).Error
	if err != nil {
		log.Printf("修改群成员状态失败: %s", err.Error())
		return
	}

	// 取出当前数据库存在的成员信息
	var oldData []entity.GroupUser
	err = client.MySQL.Model(&entity.GroupUser{}).
		Where("group_id = ?", wxId).
		Find(&oldData).Error
	if err != nil {
		log.Printf("查询群成员失败: %s", err.Error())
		return
	}
	// 将历史数据整理成map
	oldMap := make(map[string]bool)
	for _, item := range oldData {
		oldMap[item.Wxid] = item.IsMember
	}

	// 循环获取到的群成员列表
	for _, wxid := range membersIds {
		// 如果历史数据中存在，且是成员，跳过
		if isMember, ok := oldMap[wxid]; ok && isMember {
			continue
		}
		// 获取成员信息
		cp, e := utils.GetContactProfile(wxid)
		if e != nil {
			log.Printf("获取成员信息失败: %s", e.Error())
			continue
		}
		if cp.Wxid != "" {
			if _, ok := oldMap[wxid]; ok {
				// 历史数据中存在，修改
				// 修改
				gupm := map[string]any{
					"account":    cp.Account,
					"head_image": cp.HeadImage,
					"nickname":   cp.Nickname,
					"is_member":  true,
					"is_admin":   cp.Wxid == admin,
					"leave_time": nil,
				}
				err = client.MySQL.Model(&entity.GroupUser{}).Where("group_id = ?", wxId).Where("wxid = ?", wxid).Updates(gupm).Error
				if err != nil {
					log.Printf("修改群成员失败: %s", err.Error())
					continue
				}
			} else {
				// 新增的
				// 新增
				err = client.MySQL.Create(&entity.GroupUser{
					GroupId:    wxid,
					Account:    cp.Account,
					HeadImage:  cp.HeadImage,
					Nickname:   cp.Nickname,
					Wxid:       cp.Wxid,
					IsMember:   true,
					IsAdmin:    cp.Wxid == admin,
					JoinTime:   time.Now().Local(),
					LastActive: time.Now().Local(),
				}).Error
				if err != nil {
					log.Printf("新增群成员失败: %s", err.Error())
					continue
				}
			}
		}
	}
}

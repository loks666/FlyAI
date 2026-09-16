package service

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/common/applog"
	"github.com/loks666/FlyAI/common/types"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
	"time"
)

// ErrBridgeNotRunning 表示微信桥接器没有在 WECHAT_HOST 上提供服务，或者已启动但还没有
// 完成账号与窗口绑定。所有同步都依赖桥接器的 HTTP 接口，提前返回这个错误可以避免调用方
// 只看到一条 "connectex: No connection could be made" 的裸连接错误。
var ErrBridgeNotRunning = errors.New("微信桥接器未启动，请在管理页面启动桥接器后重试")

// SyncFriend
// @description: 同步好友列表
func SyncFriend() {
	_ = globalSyncQueue.do("friend:announce", func() error {
		syncFriend(true)
		return nil
	})
}

// SyncFriendCatalog refreshes contacts without sending introductions or notifications.
func SyncFriendCatalog() {
	_ = globalSyncQueue.do("friend:catalog", func() error {
		syncFriend(false)
		return nil
	})
}

func syncFriend(announce bool) {
	if !wechatBridgeReady() {
		return
	}
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
	if tx.Error != nil {
		log.Printf("开启好友同步事务失败: %s", tx.Error.Error())
		return
	}

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
			if !oldMap[item.Wxid] && announce {
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
				CustomAccount:    item.CustomAccount,
				Nickname:         item.Nickname,
				Pinyin:           item.Pinyin,
				PinyinAll:        item.PinyinAll,
				Wxid:             item.Wxid,
				IsOk:             true,
				EnableAi:         config.Conf.System.DefaultRule.Ai,
				EnableChatRank:   config.Conf.System.DefaultRule.ChatRank,
				EnableWeeklyRank: config.Conf.System.DefaultRule.ChatRank,
				EnableSummary:    config.Conf.System.DefaultRule.Summary,
				EnableWelcome:    config.Conf.System.DefaultRule.Welcome,
				EnableNews:       config.Conf.System.DefaultRule.News,
				EnableHotTop:     config.Conf.System.DefaultRule.HotTop,
				AiFreeLimit:      config.Conf.System.DefaultRule.AiFreeLimit,
				ClearMember:      0,
				LastActive:       types.DateTime(time.Now().Local()),
			}).Error
			if err != nil {
				log.Printf("新增好友失败: %s", err.Error())
				continue
			}
			if announce {
				notifyMap[item.Wxid] = item.Nickname
			}
			if conf, e := config.Conf.Resource["introduce"]; e && announce {
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
			if config.Conf.System.Domain != "" && announce {
				title := "欢迎使用微信机器人(切勿转发)"
				desc := "点我可以配置功能喔，提示非微信官方网页，点击继续访问即可"
				url := utils.GetManagerUrl(item.Wxid)
				utils.SendPublicMsg(item.Wxid, title, desc, url, 0)
			}
		}
	}

	// 通知有新成员
	if announce && len(notifyMap) > 0 && config.Conf.System.NewFriendNotify.Enable {
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
	if err = tx.Commit().Error; err != nil {
		log.Printf("提交好友同步事务失败: %s", err.Error())
		return
	}

	invalidateAllMessageIdentityCache()
	refreshFriendContactCache()
	applog.ContactSyncPrintln("同步好友列表完成")
}

// BridgeReady 供 HTTP 处理器在整段同步开始前判断桥接器是否可用。
func BridgeReady() bool {
	return wechatBridgeReady()
}

func wechatBridgeReady() bool {
	var health struct {
		Code int `json:"code"`
		Data struct {
			SendReady bool `json:"sendReady"`
		} `json:"data"`
	}
	_, err := resty.New().SetTimeout(2 * time.Second).R().SetResult(&health).Get(config.Conf.Wechat.GetURL("/health"))
	return err == nil && health.Code == 1 && health.Data.SendReady
}

// SyncGroupMembers
// @description: 同步群成员
// @param wxId
func SyncGroupMembers(wxId string) error {
	return globalSyncQueue.do("group:"+wxId, func() error {
		return syncGroupMembers(wxId)
	})
}

func syncGroupMembers(wxId string) (err error) {
	// 桥接器没起来时直接返回可读错误，不要再去 POST /api/getMemberFromChatRoom。
	if !wechatBridgeReady() {
		return ErrBridgeNotRunning
	}
	membersIds, memberNicknames, admin, err := utils.GetGroupMembers(wxId)
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
	oldMap := make(map[string]entity.GroupUser)
	for _, item := range oldData {
		oldMap[item.Wxid] = item
	}

	// 循环获取到的群成员列表
	for _, wxid := range membersIds {
		preferredNickname := memberNicknames[wxid]
		// 已存在的成员也要刷新群内昵称，避免长期保留旧的个人昵称。
		if existing, ok := oldMap[wxid]; ok && existing.IsMember {
			updates := map[string]any{"is_admin": wxid == admin}
			if preferredNickname != "" {
				updates["nickname"] = preferredNickname
			}
			if existing.WechatNickname == "" || existing.Account == "" {
				cp, profileErr := utils.GetContactProfile(wxid)
				if profileErr != nil {
					log.Printf("补充成员[%s]资料失败: %s", wxid, profileErr.Error())
				} else {
					if cp.Nickname != "" {
						updates["wechat_nickname"] = cp.Nickname
					}
					if cp.Account != "" {
						updates["account"] = cp.Account
					}
					if cp.HeadImage != "" {
						updates["head_image"] = cp.HeadImage
					}
				}
			}
			if err = client.MySQL.Model(&entity.GroupUser{}).
				Where("group_id = ?", wxId).
				Where("wxid = ?", wxid).
				Updates(updates).Error; err != nil {
				log.Printf("刷新群成员昵称失败: %s", err.Error())
			}
			continue
		}
		// 获取成员信息
		cp, e := utils.GetContactProfile(wxid)
		if e != nil {
			log.Printf("获取成员[%s]详细信息失败，使用群成员Id占位: %s", wxid, e.Error())
		}
		if cp.Wxid == "" {
			cp.Wxid = wxid
		}
		if cp.Wxid != "" {
			nickname := preferredNickname
			if nickname == "" {
				nickname = cp.Nickname
			}
			if nickname == "" {
				nickname = wxid
			}
			if _, ok := oldMap[wxid]; ok {
				// 历史数据中存在，修改
				// 修改
				gupm := map[string]any{
					"account":         cp.Account,
					"head_image":      cp.HeadImage,
					"nickname":        nickname,
					"wechat_nickname": cp.Nickname,
					"is_member":       true,
					"is_admin":        cp.Wxid == admin,
					"leave_time":      nil,
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
					GroupId:        wxId,
					Account:        cp.Account,
					HeadImage:      cp.HeadImage,
					Nickname:       nickname,
					WechatNickname: cp.Nickname,
					Wxid:           cp.Wxid,
					IsMember:       true,
					IsAdmin:        cp.Wxid == admin,
					JoinTime:       time.Now().Local(),
					LastActive:     time.Now().Local(),
				}).Error
				if err != nil {
					log.Printf("新增群成员失败: %s", err.Error())
					continue
				}
			}
		}
	}
	invalidateGroupMessageIdentityCache(wxId)
	refreshGroupMemberContactCache(wxId)
	return nil
}

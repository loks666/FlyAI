package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/dto"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/model/vo"
	"github.com/loks666/FlyAI/utils"
	"gorm.io/gorm"
	"log"
	"strings"
)

// leiGod
// @description: 雷神加速器相关接口
type leiGodI interface {
	binding(string, string, bool) string // 绑定雷神加速器账号
	info() string                        // 账户详情
	pause() string                       // 暂停加速
}

type leiGod struct {
	userId string // 用户Id
}

// newLeiGod
// @description: 创建一个雷神加速器实例
// @param userId
// @return leiGodI
func newLeiGod(userId string) leiGodI {
	return &leiGod{userId: userId}
}

// LeiGodCmd
// @description: 雷神加速器指令
// @param userId
// @param cmd
// @param args
// @return string
func LeiGodCmd(userId, cmd string, args ...string) {
	lg := newLeiGod(userId)

	var replyMsg string
	switch cmd {
	case "绑定", "b":
		var force bool
		if len(args) == 3 && args[2] == "-f" {
			force = true
		}
		replyMsg = lg.binding(args[0], args[1], force)
	case "详情", "i":
		replyMsg = lg.info()
	case "暂停", "p":
		replyMsg = lg.pause()
	default:
		replyMsg = "指令错误"
	}

	// 返回消息
	if strings.TrimSpace(replyMsg) != "" {
		utils.SendMessage(userId, "", replyMsg, 0)
	}
}

// binding
// @description: 绑定雷神加速器账号
// @receiver l
// @param account
// @param password
// @param force
// @return flag
func (l leiGod) binding(account, password string, force bool) (replyMsg string) {
	log.Printf("用户[%s]绑定雷神加速器账号[%s] -> %s", l.userId, account, password)

	// 取出已绑定的账号
	var data entity.PluginData
	client.MySQL.Where("user_id = ?", l.userId).Where("plugin_code = 'leigod'").First(&data)

	var ac vo.LeiGodAccount
	if data.UserId != "" {
		if err := json.Unmarshal([]byte(data.Data), &ac); err != nil {
			log.Printf("用户[%s]已绑定雷神账号解析失败: %v", l.userId, err)
			return
		}
		log.Printf("用户[%s]已绑定账号[%s]", l.userId, ac.Account)
	}

	// 如果已经绑定账号，且不是强制绑定，则返回
	if ac.Account != "" && !force {
		replyMsg = "您已绑定账号[" + ac.Account + "]，如需更换请使用 -f 参数: \n/雷神 绑定 账号 密码 -f"
		return
	}

	accountStr := fmt.Sprintf("{\"account\": \"%s\", \"password\":\"%s\"}", account, password)

	// 绑定账号
	var err error
	if data.UserId != "" {
		// 修改
		err = client.MySQL.Model(&data).
			Where("user_id = ?", l.userId).
			Where("plugin_code = 'leigod'").
			Update("data", accountStr).Error
	} else {
		// 新增
		data = entity.PluginData{
			UserId:     l.userId,
			PluginCode: "leigod",
			Data:       accountStr,
		}
		err = client.MySQL.Create(&data).Error
	}

	if err != nil {
		log.Printf("用户[%s]绑定雷神账号失败: %v", l.userId, err)
		replyMsg = "绑定失败: " + err.Error()
	} else {
		replyMsg = "绑定成功"
	}

	return
}

// info
// @description: 账户详情
// @receiver l
// @return replyMsg
func (l leiGod) info() (replyMsg string) {
	log.Printf("用户[%s]获取雷神账户详情", l.userId)

	// 取出已绑定的账号
	var data entity.PluginData
	err := client.MySQL.Where("user_id = ?", l.userId).Where("plugin_code = 'leigod'").First(&data).Error
	if err != nil {
		log.Printf("用户[%s]获取雷神账户详情失败: %v", l.userId, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			replyMsg = "您还未绑定账号，请先绑定后再使用，绑定指定:\n/雷神 绑定 你的账号 你的密码"
		} else {
			replyMsg = "系统错误: " + err.Error()
		}
		return
	}

	// 解析为结构体
	var ac vo.LeiGodAccount
	if err = json.Unmarshal([]byte(data.Data), &ac); err != nil {
		log.Printf("用户[%s]已绑定雷神账号解析失败: %v", l.userId, err)
		replyMsg = "系统炸了，请耐心等待修复"
		return
	}

	lgu := utils.LeiGodUtil(ac.Account, ac.Password)
	if err = lgu.Login(); err != nil {
		return "登录失败: " + err.Error()
	}
	var ui dto.LeiGodUserInfoResp
	if ui, err = lgu.Info(); err != nil {
		return "获取详情失败: " + err.Error()
	}
	replyMsg = fmt.Sprintf("#账户 %s\n#剩余时长 %s\n#暂停状态 %s\n#最后暂停时间 %s",
		ui.Mobile, ui.ExpiryTime, ui.PauseStatus, ui.LastPauseTime)
	return
}

// pause
// @description: 暂停加速
// @receiver l
// @return flag
func (l leiGod) pause() (replyMsg string) {
	log.Printf("用户[%s]暂停加速", l.userId)

	// 取出已绑定的账号
	var data entity.PluginData
	err := client.MySQL.Where("user_id = ?", l.userId).Where("plugin_code = 'leigod'").First(&data).Error
	if err != nil {
		log.Printf("用户[%s]获取雷神账户详情失败: %v", l.userId, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			replyMsg = "您还未绑定账号，请先绑定后再使用，绑定指定:\n/雷神 绑定 你的账号 你的密码"
		} else {
			replyMsg = "系统错误: " + err.Error()
		}
		return
	}

	// 解析为结构体
	var ac vo.LeiGodAccount
	if err = json.Unmarshal([]byte(data.Data), &ac); err != nil {
		log.Printf("用户[%s]已绑定雷神账号解析失败: %v", l.userId, err)
		replyMsg = "系统炸了，请耐心等待修复"
		return
	}

	lgu := utils.LeiGodUtil(ac.Account, ac.Password)
	if err = lgu.Login(); err != nil {
		return "登录失败: " + err.Error()
	}
	if err = lgu.Pause(); err != nil {
		return "暂停失败: " + err.Error()
	}

	return "暂停成功"
}

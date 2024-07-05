package plugins

import (
	"go-wechat/client"
	"go-wechat/config"
	"go-wechat/model/entity"
	"go-wechat/plugin"
	"go-wechat/utils"
)

// WelcomeNew
// @description: 欢迎新成员
// @param m
func WelcomeNew(m *plugin.MessageContext) {
	// 判断是否开启迎新
	var count int64
	client.MySQL.Model(&entity.Friend{}).Where("enable_welcome IS TRUE").Where("wxid = ?", m.FromUser).Count(&count)
	if count < 1 {
		return
	}

	// 读取欢迎新成员配置
	conf, ok := config.Conf.Resource["welcome-new"]
	if !ok {
		// 未配置，跳过
		return
	}
	switch conf.Type {
	case "text":
		// 文字类型
		utils.SendMessage(m.FromUser, "", conf.Path, 0)
	case "image":
		// 图片类型
		utils.SendImage(m.FromUser, conf.Path, 0)
	case "emotion":
		// 表情类型
		utils.SendEmotion(m.FromUser, conf.Path, 0)
	}
}

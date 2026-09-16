package service

import (
	"errors"
	"strings"
	"unicode"

	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
)

var ErrInvalidRobotWechatAccount = errors.New("请输入有效的机器人微信号")

const bridgeSettingID uint = 1

func normalizeRobotWechatAccount(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidRobotWechatAccount
	}
	return value, nil
}

func GetBridgeSetting() (entity.BridgeSetting, error) {
	setting := entity.BridgeSetting{Id: bridgeSettingID}
	err := client.MySQL.FirstOrCreate(&setting, entity.BridgeSetting{Id: bridgeSettingID}).Error
	return setting, err
}

func SaveBridgeWechatAccount(value string) (entity.BridgeSetting, error) {
	account, err := normalizeRobotWechatAccount(value)
	if err != nil {
		return entity.BridgeSetting{}, err
	}
	setting, err := GetBridgeSetting()
	if err != nil {
		return entity.BridgeSetting{}, err
	}
	if !strings.EqualFold(setting.WechatAccount, account) {
		setting.ResolvedWxid = ""
		setting.AccountDirectory = ""
	}
	setting.WechatAccount = account
	err = client.MySQL.Save(&setting).Error
	return setting, err
}

func SaveBridgeResolution(wxid, accountDirectory string) error {
	setting, err := GetBridgeSetting()
	if err != nil {
		return err
	}
	setting.ResolvedWxid = strings.TrimSpace(wxid)
	setting.AccountDirectory = strings.TrimSpace(accountDirectory)
	return client.MySQL.Save(&setting).Error
}

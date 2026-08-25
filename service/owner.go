package service

import (
	"errors"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"gorm.io/gorm"
	"strings"
	"unicode"
)

var (
	ErrInvalidWechatAccount = errors.New("请输入有效的微信号")
	ErrDirectWxidNotAllowed = errors.New("这里只支持普通微信号，不支持直接填写 wxid")
	ErrOwnerContactNotFound = errors.New("通讯录中未找到该微信号，请先添加为机器人好友并同步通讯录")
	ErrOwnerAlreadyExists   = errors.New("该微信号已经绑定")
	ErrOwnerNotFound        = errors.New("主人配置不存在")
)

func normalizeWechatAccount(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidWechatAccount
	}
	if strings.HasPrefix(strings.ToLower(value), "wxid_") {
		return "", ErrDirectWxidNotAllowed
	}
	return value, nil
}

func resolveOwnerContact(wechatAccount string) (entity.Friend, error) {
	var contact entity.Friend
	err := client.MySQL.
		Where("custom_account = ?", wechatAccount).
		Where("wxid NOT LIKE '%@chatroom'").
		Where("is_ok IS TRUE").
		First(&contact).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return contact, ErrOwnerContactNotFound
	}
	return contact, err
}

func GetOwners() (owners []entity.Owner, err error) {
	err = client.MySQL.Order("id ASC").Find(&owners).Error
	return
}

func CreateOwner(wechatAccount string) (entity.Owner, error) {
	account, err := normalizeWechatAccount(wechatAccount)
	if err != nil {
		return entity.Owner{}, err
	}
	contact, err := resolveOwnerContact(account)
	if err != nil {
		return entity.Owner{}, err
	}
	var count int64
	if err = client.MySQL.Model(&entity.Owner{}).
		Where("wechat_account = ? OR wxid = ?", contact.CustomAccount, contact.Wxid).
		Count(&count).Error; err != nil {
		return entity.Owner{}, err
	}
	if count > 0 {
		return entity.Owner{}, ErrOwnerAlreadyExists
	}
	owner := entity.Owner{WechatAccount: contact.CustomAccount, Wxid: contact.Wxid, Nickname: contact.Nickname}
	err = client.MySQL.Create(&owner).Error
	return owner, err
}

func UpdateOwner(id uint, wechatAccount string) (entity.Owner, error) {
	account, err := normalizeWechatAccount(wechatAccount)
	if err != nil {
		return entity.Owner{}, err
	}
	var owner entity.Owner
	if err = client.MySQL.First(&owner, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.Owner{}, ErrOwnerNotFound
	} else if err != nil {
		return entity.Owner{}, err
	}
	contact, err := resolveOwnerContact(account)
	if err != nil {
		return entity.Owner{}, err
	}
	var count int64
	if err = client.MySQL.Model(&entity.Owner{}).
		Where("id <> ?", id).
		Where("wechat_account = ? OR wxid = ?", contact.CustomAccount, contact.Wxid).
		Count(&count).Error; err != nil {
		return entity.Owner{}, err
	}
	if count > 0 {
		return entity.Owner{}, ErrOwnerAlreadyExists
	}
	owner.WechatAccount = contact.CustomAccount
	owner.Wxid = contact.Wxid
	owner.Nickname = contact.Nickname
	err = client.MySQL.Save(&owner).Error
	return owner, err
}

func DeleteOwner(id uint) error {
	result := client.MySQL.Delete(&entity.Owner{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOwnerNotFound
	}
	return nil
}

func IsOwnerWxid(wxid string) bool {
	if wxid == "" {
		return false
	}
	var count int64
	if err := client.MySQL.Model(&entity.Owner{}).Where("wxid = ?", wxid).Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

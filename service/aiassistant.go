package service

import (
	"go-wechat/client"
	"go-wechat/entity"
)

// GetAllAiAssistant
// @description: 取出所有AI助手
// @return records
func GetAllAiAssistant() (records []entity.AiAssistant, err error) {
	err = client.MySQL.Order("created_at DESC").Find(&records).Error
	return
}

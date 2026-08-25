package service

import (
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
)

// GetAllAiAssistant
// @description: 取出所有AI助手
// @return records
func GetAllAiAssistant() (records []entity.AiAssistant, err error) {
	err = client.MySQL.Order("created_at DESC").Find(&records).Error
	return
}

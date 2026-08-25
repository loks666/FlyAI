package entity

import (
	"github.com/google/uuid"
	"github.com/loks666/FlyAI/common/types"
	"gorm.io/gorm"
	"strings"
)

// AiAssistant
// @description: AI助手表
type AiAssistant struct {
	Id          string         `json:"id" gorm:"type:varchar(32);primarykey"`
	CreatedAt   types.DateTime `json:"createdAt"`
	Name        string         `json:"name" gorm:"type:varchar(10);not null;comment:'名称'"`
	Personality string         `json:"personality" gorm:"type:varchar(999);not null;comment:'人设'"`
	Model       string         `json:"dto" gorm:"type:varchar(50);not null;comment:'使用的模型'"`
	Enable      bool           `json:"enable" gorm:"type:tinyint(1);not null;default:1;comment:'是否启用'"`
}

// TableName
// @description: 表名
// @receiver AiAssistant
// @return string
func (AiAssistant) TableName() string {
	return "t_ai_assistant"
}

// BeforeCreate
// @description: 创建数据库对象之前生成UUID
// @receiver m
// @param *gorm.DB
// @return err
func (m *AiAssistant) BeforeCreate(*gorm.DB) (err error) {
	if m.Id == "" {
		m.Id = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	return
}

package entity

import (
	"github.com/google/uuid"
	"go-wechat/common/types"
	"gorm.io/gorm"
	"strings"
)

// HotTop
// @description: 热榜数据
type HotTop struct {
	Id        string         `json:"id" gorm:"type:varchar(32);primarykey"`
	CreatedAt types.DateTime `json:"createdAt"`
	Channel   string         `json:"channel"`   // 渠道
	Title     string         `json:"title"`     // 标题
	Hot       string         `json:"hot"`       // 热度
	Url       string         `json:"url"`       // 链接
	MobileUrl string         `json:"mobileUrl"` // 手机端链接
}

// TableName
// @description: 表名
// @receiver HotTop
// @return string
func (HotTop) TableName() string {
	return "t_hot_top"
}

// BeforeCreate
// @description: 创建数据库对象之前生成UUID
// @receiver m
// @param *gorm.DB
// @return err
func (m *HotTop) BeforeCreate(*gorm.DB) (err error) {
	if m.Id == "" {
		m.Id = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	return
}

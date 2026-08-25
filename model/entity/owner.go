package entity

import "time"

// Owner is a trusted WeChat contact allowed to run owner-only group commands.
type Owner struct {
	Id            uint      `json:"id" gorm:"primaryKey"`
	WechatAccount string    `json:"wechatAccount" gorm:"size:128;uniqueIndex;not null"`
	Wxid          string    `json:"wxid" gorm:"size:128;uniqueIndex;not null"`
	Nickname      string    `json:"nickname" gorm:"size:255;not null"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Owner) TableName() string {
	return "t_owner"
}

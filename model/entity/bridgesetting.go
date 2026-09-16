package entity

import "time"

// BridgeSetting stores the single local WeChat account selected for FlyAI.
// Runtime process/window state is deliberately not persisted because it can
// become stale as soon as WeChat or FlyAI exits.
type BridgeSetting struct {
	Id               uint      `json:"id" gorm:"primaryKey"`
	WechatAccount    string    `json:"wechatAccount" gorm:"size:128;not null;default:''"`
	ResolvedWxid     string    `json:"resolvedWxid" gorm:"size:128;not null;default:''"`
	AccountDirectory string    `json:"accountDirectory" gorm:"size:1024;not null;default:''"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (BridgeSetting) TableName() string {
	return "t_bridge_setting"
}

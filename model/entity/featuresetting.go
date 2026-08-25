package entity

import "time"

// FeatureSetting stores global scheduling state for one extensible feature.
type FeatureSetting struct {
	Feature          string    `json:"feature" gorm:"primaryKey;size:32"`
	SendTime         string    `json:"sendTime" gorm:"size:5;not null"`
	Schedule         string    `json:"schedule" gorm:"size:16;not null;default:'daily'"`
	MessageThreshold int       `json:"messageThreshold" gorm:"not null;default:0"`
	LastRunDate      string    `json:"lastRunDate" gorm:"size:10;not null;default:''"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (FeatureSetting) TableName() string {
	return "t_feature_setting"
}

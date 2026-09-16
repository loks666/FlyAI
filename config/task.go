package config

// task
// @description: 定时任务
type task struct {
	Enable       bool        `json:"enable" yaml:"enable"`             // 是否启用
	News         syncFriends `json:"news" yaml:"news"`                 // 每日早报
	SyncFriends  syncFriends `json:"syncFriends" yaml:"syncFriends"`   // 同步好友
	GroupSummary syncFriends `json:"groupSummary" yaml:"groupSummary"` // 群聊总结
	HotTop       syncFriends `json:"hotTop" yaml:"hotTop"`             // 热搜排行榜
}

// syncFriends
// @description: 同步好友
type syncFriends struct {
	Enable bool   `json:"enable" yaml:"enable"` // 是否启用
	Cron   string `json:"cron" yaml:"cron"`     // 定时任务表达式
}

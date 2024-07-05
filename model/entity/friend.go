package entity

import (
	"time"
)

// Friend
// @description: 好友列表
type Friend struct {
	Wxid           string    `json:"wxid"`                                                     // 微信原始Id
	CustomAccount  string    `json:"customAccount"`                                            // 微信号
	Nickname       string    `json:"nickname"`                                                 // 昵称
	Pinyin         string    `json:"pinyin"`                                                   // 昵称拼音大写首字母
	PinyinAll      string    `json:"pinyinAll"`                                                // 昵称全拼
	LastActive     time.Time `json:"lastActive"`                                               // 最后活跃时间
	EnableAi       bool      `json:"enableAI" gorm:"type:tinyint(1) default 0 not null"`       // 是否使用AI
	AiModel        string    `json:"aiModel"`                                                  // AI模型
	Prompt         string    `json:"prompt"`                                                   // 提示词
	EnableChatRank bool      `json:"enableChatRank" gorm:"type:tinyint(1) default 0 not null"` // 是否使用聊天排行
	EnableWelcome  bool      `json:"enableWelcome" gorm:"type:tinyint(1) default 0 not null"`  // 是否启用迎新
	EnableSummary  bool      `json:"enableSummary" gorm:"type:tinyint(1) default 0 not null"`  // 是否启用总结
	EnableNews     bool      `json:"enableNews" gorm:"type:tinyint(1) default 0 not null"`     // 是否启用新闻
	ClearMember    int       `json:"clearMember"`                                              // 清理成员配置（多少天未活跃的）
	IsOk           bool      `json:"isOk" gorm:"type:tinyint(1) default 0 not null"`           // 是否正常
}

func (Friend) TableName() string {
	return "t_friend"
}

// GroupUser
// @description: 群成员
type GroupUser struct {
	GroupId      string     `json:"groupId"`                                                // 群Id
	Wxid         string     `json:"wxid"`                                                   // 微信Id
	Account      string     `json:"account"`                                                // 账号
	HeadImage    string     `json:"headImage"`                                              // 头像
	Nickname     string     `json:"nickname"`                                               // 昵称
	IsMember     bool       `json:"isMember" gorm:"type:tinyint(1) default 0 not null"`     // 是否群成员
	IsAdmin      bool       `json:"isAdmin" gorm:"type:tinyint(1) default 0 not null"`      // 是否群主
	JoinTime     time.Time  `json:"joinTime"`                                               // 加入时间
	LastActive   time.Time  `json:"lastActive"`                                             // 最后活跃时间
	LeaveTime    *time.Time `json:"leaveTime"`                                              // 离开时间
	SkipChatRank bool       `json:"skipChatRank" gorm:"type:tinyint(1) default 0 not null"` // 是否跳过聊天排行
}

func (GroupUser) TableName() string {
	return "t_group_user"
}

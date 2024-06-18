package vo

import (
	"go-wechat/common/types"
)

// FriendItem
// @description: 好友列表数据
type FriendItem struct {
	CustomAccount  string         // 微信号
	Nickname       string         // 昵称
	Pinyin         string         // 昵称拼音大写首字母
	PinyinAll      string         // 昵称全拼
	Wxid           string         // 微信原始Id
	LastActive     types.DateTime // 最后活跃时间
	EnableAi       bool           // 是否使用AI
	AiModel        string         // AI模型
	Prompt         string         // AI助手或者自定义提示词
	EnableChatRank bool           // 是否使用聊天排行
	EnableWelcome  bool           // 是否使用迎新
	EnableCommand  bool           // 是否启用指令
	EnableSummary  bool           // 是否启用总结
	EnableNews     bool           // 是否启用新闻
	ClearMember    int            // 清理成员配置（多少天未活跃的）
	IsOk           bool           // 是否还在通讯库(群聊是要还在群里也算)
}

// GroupUserItem
// @description: 群成员列表数据
type GroupUserItem struct {
	Wxid         string         `json:"wxid"`          // 微信Id
	Account      string         `json:"account"`       // 账号
	HeadImage    string         `json:"headImage"`     // 头像
	Nickname     string         `json:"nickname"`      // 昵称
	IsMember     bool           `json:"isMember" `     // 是否群成员
	IsAdmin      bool           `json:"isAdmin"`       // 是否群主
	JoinTime     types.DateTime `json:"joinTime"`      // 加入时间
	LastActive   types.DateTime `json:"lastActive"`    // 最后活跃时间
	LeaveTime    types.DateTime `json:"leaveTime"`     // 离开时间
	SkipChatRank bool           `json:"skipChatRank" ` // 是否跳过聊天排行
}

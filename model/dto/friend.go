package dto

// FriendItem
// @description: 好友列表数据
type FriendItem struct {
	CustomAccount string `json:"customAccount"` // 微信号
	EncryptName   string `json:"encryptName"`   // 不知道
	Nickname      string `json:"nickname"`      // 昵称
	Pinyin        string `json:"pinyin"`        // 昵称拼音大写首字母
	PinyinAll     string `json:"pinyinAll"`     // 昵称全拼
	Reserved1     int    `json:"reserved1"`     // 未知
	Reserved2     int    `json:"reserved2"`     // 未知
	Type          int    `json:"type"`          // 类型
	VerifyFlag    int    `json:"verifyFlag"`    // 未知
	Wxid          string `json:"wxid"`          // 微信原始Id
}

// GroupUser
// @description: 群成员返回结果
type GroupUser struct {
	Admin          string `json:"admin"`          // 群主微信
	AdminNickname  string `json:"adminNickname"`  // 群主昵称
	ChatRoomId     string `json:"chatRoomId"`     // 群Id
	MemberNickname string `json:"memberNickname"` // 成员昵称 `^G`切割
	Members        string `json:"members"`        // 成员Id `^G`切割
}

type GroupActiveMembers struct {
	ChatRoomId string           `json:"chatRoomId"`
	Members    []string         `json:"members"`
	Counts     map[string]int64 `json:"counts"`
}

// ContactProfile
// @description: 好友资料
type ContactProfile struct {
	Account   string `json:"account"`   // 账号
	HeadImage string `json:"headImage"` // 头像
	Nickname  string `json:"nickname"`  // 昵称
	V3        string `json:"v3"`        // v3
	Wxid      string `json:"wxid"`      // 微信Id
}

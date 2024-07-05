package dto

// RobotUserInfo
// @description: 机器人用户信息
type RobotUserInfo struct {
	WxId            string `json:"wxid"`            // 微信Id
	Account         string `json:"account"`         // 微信号
	Name            string `json:"name"`            // 昵称
	HeadImage       string `json:"headImage"`       // 头像
	Mobile          string `json:"mobile"`          // 手机
	Signature       string `json:"signature"`       // 个人签名
	Country         string `json:"country"`         // 国家
	Province        string `json:"province"`        // 省
	City            string `json:"city"`            // 城市
	CurrentDataPath string `json:"currentDataPath"` // 当前数据目录,登录的账号目录
	DataSavePath    string `json:"dataSavePath"`    // 微信保存目录
	DbKey           string `json:"dbKey"`           // 数据库的SQLCipher的加密key，可以使用该key配合decrypt.py解密数据库
}

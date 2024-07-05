package dto

// MorningPost
// @description: 每日早报返回结构体
type MorningPost struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Date      string   `json:"date"`       // 新闻日期
		News      []string `json:"news"`       // 新闻标题文字版
		WeiYu     string   `json:"weiyu"`      // 微语，就是一句屁话
		Image     string   `json:"image"`      // 早报完整图片
		HeadImage string   `json:"head_image"` // 早报头部图片
	} `json:"data"`
	Time  int    `json:"time"`
	Usage int    `json:"usage"`
	LogId string `json:"log_id"`
}

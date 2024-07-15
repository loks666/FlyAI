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

// HotTop
// @description: 热搜排行榜返回结构体
type HotTop struct {
	Success    bool             `json:"success"`     // 是否成功
	Name       string           `json:"name"`        // 渠道
	Subtitle   string           `json:"subtitle"`    // 副标题
	UpdateTime string           `json:"update_time"` // 更新时间
	Data       []HotTopDataItem `json:"data"`        // 数据
}

// HotTopDataItem
// @description: 热搜排行榜数据项
type HotTopDataItem struct {
	Index    int    `json:"index"`    // 排行
	Title    string `json:"title"`    // 标题
	Hot      string `json:"hot"`      // 热度
	Url      string `json:"url"`      // 链接
	MobilUrl string `json:"mobilUrl"` // 手机端链接
}

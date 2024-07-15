package utils

import (
	"github.com/go-resty/resty/v2"
	"go-wechat/config"
	"go-wechat/model/dto"
	"log"
)

// News
// @description: 新闻
type News interface {
	MorningPost() []string           // 早报
	GetHotTop() []dto.HotTopDataItem // 获取热搜排行榜
}

type news struct{}

// NewsUtil
// @description: 新闻工具
// @param account
// @param password
// @return LeiGod
func NewsUtil() News {
	return &news{}
}

// MorningPost
// @description: 早报
// @receiver news
// @return records
func (news) MorningPost() (records []string) {
	var newsResp dto.MorningPost

	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetQueryParam("token", config.Conf.System.AlApiToken).
		SetResult(&newsResp).
		Post("https://v2.alapi.cn/api/zaobao")
	if err != nil {
		log.Printf("每日早报获取失败: %s", err.Error())
		return
	}
	log.Printf("每日早报获取结果: %s", unicodeToText(resp.String()))

	records = newsResp.Data.News
	return
}

// GetHotTop
// @description: 获取热搜排行榜
// @receiver news
// @return records
func (news) GetHotTop() (records []dto.HotTopDataItem) {
	var respData dto.HotTop
	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetResult(&respData).
		Get("https://api.vvhan.com/api/hotlist/baiduRD")
	if err != nil {
		log.Printf("百度热榜获取失败: %s", err.Error())
		return
	}
	log.Printf("百度热榜获取结果: %s", unicodeToText(resp.String()))
	if !respData.Success {
		log.Println("百度热榜获取失败")
		return
	}
	return respData.Data
}

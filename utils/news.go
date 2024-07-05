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
	MorningPost() []string // 早报
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
		log.Panicf("每日早报获取失败: %s", err.Error())
	}
	log.Printf("每日早报获取结果: %s", unicodeToText(resp.String()))

	records = newsResp.Data.News
	return
}

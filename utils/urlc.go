package utils

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"go-wechat/config"
	"go-wechat/model/dto"
	"log"
)

// GenShortUrl
// @description: 生成短链接
// @param url
// @return shortUrl
func GenShortUrl(url string) (shortUrl string) {
	// 组装参数
	param := map[string]any{
		"url": url,
	}
	pbs, _ := json.Marshal(param)

	var respData dto.ShortUrlResponse
	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetAuthScheme("Token").
		SetAuthToken(config.Conf.System.UrlcApiToken).
		SetBody(string(pbs)).
		Post("https://www.urlc.cn/api/url/add")
	if err != nil {
		log.Printf("短链接获取失败: %s", err.Error())
		return
	}
	log.Printf("短链接获取结果: %s", resp.String())
	_ = json.Unmarshal(resp.Body(), &respData)
	return respData.Short
}

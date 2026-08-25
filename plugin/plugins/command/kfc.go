package command

import (
	"github.com/go-resty/resty/v2"
	"github.com/loks666/FlyAI/utils"
	"log"
	"net/http"
)

// KfcCrazyThursdayCmd
// @description: 肯德基疯狂星期四文案
// @param userId string 发信人
func KfcCrazyThursdayCmd(userId string) {
	// 随机选一个接口调用
	str := kfcApi1()
	if str == "" {
		str = kfcApi2()
	}
	if str == "" {
		str = kfcApi3()
	}
	if str == "" {
		str = "文案获取失败"
	}

	// 发送消息
	utils.SendMessage(userId, "", str, 0)
}

// kfcApi1
// @description: 肯德基疯狂星期四文案接口1
// @return string
func kfcApi1() string {
	res := resty.New()
	resp, err := res.R().
		Post("https://api.jixs.cc/api/wenan-fkxqs/index.php")
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("KFC接口1文案获取失败: %v", err)
		return ""
	}
	log.Printf("KFC接口1文案获取结果: %s", resp.String())
	return resp.String()
}

// kfcApi2
// @description: 肯德基疯狂星期四文案接口2
// @return string
func kfcApi2() string {
	type result struct {
		Code int    `json:"code"`
		Text string `json:"text"`
		Data struct {
			Msg string `json:"msg"`
		} `json:"data"`
	}

	var resData result

	res := resty.New()
	resp, err := res.R().
		SetResult(&resData).
		Post("https://api.jixs.cc/api/wenan-fkxqs/index.php")
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("KFC接口2文案获取失败: %v", err)
		return ""
	}
	log.Printf("KFC接口2文案获取结果: %s", resp.String())
	if resData.Data.Msg != "" {
		return resData.Data.Msg
	}
	return resp.String()
}

// kfcApi3
// @description: 肯德基疯狂星期四文案接口3
// @return string
func kfcApi3() string {
	type result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Text string `json:"text"`
	}

	var resData result

	res := resty.New()
	resp, err := res.R().
		SetResult(&resData).
		Post("https://api.pearktrue.cn/api/kfc")
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("KFC接口3文案获取失败: %v", err)
		return ""
	}
	log.Printf("KFC接口3文案获取结果: %s", resp.String())
	if resData.Text != "" {
		return resData.Text
	}
	return resp.String()
}

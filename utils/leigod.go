package utils

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go-wechat/model/dto"
	"log"
)

// LeiGod
// @description: 雷神加速器相关接口
type LeiGod interface {
	Login() error                          // 登录
	Info() (dto.LeiGodUserInfoResp, error) // 获取用户信息
	Pause() error                          // 暂停加速
}

type leiGod struct {
	account, password string // 账号、密码
	token             string
}

// LeiGodUtil
// @description: 创建一个雷神加速器工具类
// @param userId
// @return leiGodI
func LeiGodUtil(account, password string) LeiGod {
	// 把密码md5一下
	hash := md5.New()
	hash.Write([]byte(password))
	password = fmt.Sprintf("%x", hash.Sum(nil))

	return &leiGod{account: account, password: password}
}

// Login
// @description: 登录
// @receiver l
// @return string
func (l *leiGod) Login() (err error) {
	// 组装参数
	param := map[string]any{
		"account_token": nil,
		"country_code":  86,
		"lang":          "zh_CN",
		"os_type":       4,
		"mobile_num":    l.account,
		"username":      l.account,
		"password":      l.password,
		"region_code":   1,
		"src_channel":   "guanwang",
		"sem_ad_img_url": map[string]any{
			"btn_yrl": "",
			"url":     "",
		},
	}
	pbs, _ := json.Marshal(param)

	var loginResp dto.Response[any]
	var resp *resty.Response

	res := resty.New()
	resp, err = res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		SetResult(&loginResp).
		Post("https://webapi.leigod.com/api/auth/login")
	if err != nil {
		log.Panicf("雷神加速器登录失败: %s", err.Error())
	}
	log.Printf("雷神加速器登录结果: %s", unicodeToText(resp.String()))

	// 返回状态码不是0表示有错
	if loginResp.Code != 0 {
		return errors.New(loginResp.Msg)
	}

	// 将Data字段转为结构体
	var bs []byte
	if bs, err = json.Marshal(loginResp.Data); err != nil {
		return
	}

	var loginInfo dto.LeiGodLoginResp
	if err = json.Unmarshal(bs, &loginInfo); err != nil {
		return
	}

	if loginInfo.LoginInfo.AccountToken != "" {
		l.token = loginInfo.LoginInfo.AccountToken
	}

	return
}

// Info
// @description: 获取用户信息
// @receiver l
// @return string
func (l *leiGod) Info() (ui dto.LeiGodUserInfoResp, err error) {
	// 组装参数
	param := map[string]any{
		"account_token": l.token,
		"lang":          "zh_CN",
		"os_type":       4,
	}
	pbs, _ := json.Marshal(param)

	var userInfoResp dto.Response[dto.LeiGodUserInfoResp]
	var resp *resty.Response

	res := resty.New()
	resp, err = res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		SetResult(&userInfoResp).
		Post("https://webapi.leigod.com/api/user/info")
	if err != nil {
		log.Panicf("雷神加速器用户信息获取失败: %s", err.Error())
	}
	log.Printf("雷神加速器用户信息获取结果: %s", unicodeToText(resp.String()))

	// 返回状态码不是0表示有错
	if userInfoResp.Code != 0 {
		err = errors.New(userInfoResp.Msg)
		return
	}

	return userInfoResp.Data, err
}

// Pause
// @description: 暂停加速
// @receiver l
// @return string
func (l *leiGod) Pause() (err error) {
	// 组装参数
	param := map[string]any{
		"account_token": l.token,
		"lang":          "zh_CN",
		"os_type":       4,
	}
	pbs, _ := json.Marshal(param)

	var pauseResp dto.Response[any]
	var resp *resty.Response

	res := resty.New()
	resp, err = res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		SetResult(&pauseResp).
		Post("https://webapi.leigod.com/api/user/pause")
	if err != nil {
		log.Panicf("雷神加速器暂停失败: %s", err.Error())
	}
	log.Printf("雷神加速器暂停结果: %s", unicodeToText(resp.String()))

	// 返回状态码不是0表示有错
	if pauseResp.Code != 0 {
		err = errors.New(pauseResp.Msg)
		return
	}

	return
}

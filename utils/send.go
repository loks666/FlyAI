package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/model/dto"
	"log"
	"os"
	"strconv"
	"time"
)

// SendMessage
// @description: 发送消息
// @param toId
// @param atId
// @param msg
func SendMessage(toId, atId, msg string, retryCount int) (err error) {
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SEND")); flag {
		return
	}

	if retryCount > 5 {
		log.Printf("重试五次失败，停止发送")
		err = errors.New("重试五次失败，停止发送")
		return
	}

	// 组装参数
	param := map[string]any{
		"wxid": toId, // 群或好友Id
		"msg":  msg,  // 消息
	}

	// 接口地址
	apiUrl := config.Conf.Wechat.GetURL("/api/sendTextMsg")
	if atId != "" {
		apiUrl = config.Conf.Wechat.GetURL("/api/sendAtText")
		param = map[string]any{
			"chatRoomId": toId,
			"wxids":      atId,
			"msg":        msg, // 消息
		}
	}
	pbs, _ := json.Marshal(param)

	var result dto.Response[struct {
		Sent   bool   `json:"sent"`
		Detail string `json:"detail"`
	}]
	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetResult(&result).
		SetBody(string(pbs)).
		Post(apiUrl)
	if err != nil {
		log.Printf("发送文本消息失败: %s", err.Error())
		// 休眠五秒后重新发送
		time.Sleep(5 * time.Second)
		return SendMessage(toId, atId, msg, retryCount+1)
	}
	log.Printf("发送文本消息结果: %s", resp.String())
	if resp.IsError() {
		return fmt.Errorf("微信桥接服务返回 HTTP %d", resp.StatusCode())
	}
	if result.Code != 1 || !result.Data.Sent {
		detail := result.Data.Detail
		if detail == "" {
			detail = result.Msg
		}
		if detail == "" {
			detail = "微信桥接服务未确认发送成功"
		}
		return errors.New(detail)
	}
	return
}

// SendImage
// @description: 发送图片
// @param toId string 群或者好友Id
// @param imgPath string 图片路径
// @param retryCount int 重试次数
func SendImage(toId, imgPath string, retryCount int) {
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SEND")); flag {
		return
	}

	if retryCount > 5 {
		log.Printf("重试五次失败，停止发送")
		return
	}

	// 组装参数
	param := map[string]any{
		"wxid":      toId,    // 群或好友Id
		"imagePath": imgPath, // 图片地址
	}
	pbs, _ := json.Marshal(param)

	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		Post(config.Conf.Wechat.GetURL("/api/sendImagesMsg"))
	if err != nil {
		log.Printf("发送图片消息失败: %s", err.Error())
		// 休眠五秒后重新发送
		time.Sleep(5 * time.Second)
		SendImage(toId, imgPath, retryCount+1)
	}
	log.Printf("发送图片消息结果: %s", resp.String())
}

// SendEmotion
// @description: 发送自定义表情包
// @param toId string 群或者好友Id
// @param emotionHash string 表情包hash(md5值)
// @param retryCount int 重试次数
func SendEmotion(toId, emotionHash string, retryCount int) {
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SEND")); flag {
		return
	}

	if retryCount > 5 {
		log.Printf("重试五次失败，停止发送")
		return
	}

	// 组装表情包本地地址
	// 规则：机器人数据目录\FileStorage\CustomEmotion\表情包hash前两位\表情包hash
	emotionPath := fmt.Sprintf("%sFileStorage\\CustomEmotion\\%s\\%s",
		current.GetRobotInfo().CurrentDataPath, emotionHash[:2], emotionHash)

	// 组装参数
	param := map[string]any{
		"wxid":     toId,        // 群或好友Id
		"filePath": emotionPath, // 图片地址
	}
	pbs, _ := json.Marshal(param)

	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		Post(config.Conf.Wechat.GetURL("/api/sendCustomEmotion"))
	if err != nil {
		log.Printf("发送表情包消息失败: %s", err.Error())
		// 休眠五秒后重新发送
		time.Sleep(5 * time.Second)
		SendImage(toId, emotionHash, retryCount+1)
	}
	log.Printf("发送表情包消息结果: %s", resp.String())
}

// DeleteGroupMember
// @description: 删除群成员
// @param chatRoomId 群Id
// @param memberIds 成员id,用','分隔
// @param retryCount 重试次数
// @param isSure 是否确认删除
func DeleteGroupMember(chatRoomId, memberIds string, retryCount int, isSure bool) {
	if retryCount > 5 {
		log.Printf("重试五次失败，停止发送")
		return
	}

	// 组装参数
	param := map[string]any{
		"chatRoomId": chatRoomId, // 群Id
		"memberIds":  memberIds,  // 成员id
	}
	pbs, _ := json.Marshal(param)

	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		Post(config.Conf.Wechat.GetURL("/api/delMemberFromChatRoom"))
	if err != nil {
		log.Printf("删除群成员失败: %s", err.Error())
		// 休眠五秒后重新发送
		time.Sleep(5 * time.Second)
		DeleteGroupMember(chatRoomId, memberIds, retryCount+1, isSure)
	}
	log.Printf("[%s]删除群成员结果: %s", chatRoomId, resp.String())
	// 这个逼接口要调用两次，第一次调用成功，第二次调用才会真正删除
	if !isSure {
		DeleteGroupMember(chatRoomId, memberIds, 5, true)
	}
}

// QuitChatroom
// @description: 退出群聊
// @param chatRoomId string 群Id
// @param retryCount int 重试次数
func QuitChatroom(chatRoomId string, retryCount int) {
	if retryCount > 5 {
		log.Printf("重试五次失败，停止发送")
		return
	}

	// 组装参数
	param := map[string]any{
		"chatRoomId": chatRoomId, // 群Id
	}
	pbs, _ := json.Marshal(param)

	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		Post(config.Conf.Wechat.GetURL("/api/quitChatRoom"))
	if err != nil {
		log.Printf("退群失败: %s", err.Error())
		// 休眠五秒后重新发送
		time.Sleep(5 * time.Second)
		QuitChatroom(chatRoomId, retryCount+1)
	}
	log.Printf("退群结果: %s", resp.String())
}

// SendPublicMsg
// @description: 发送公众号消息
// @param wxId string 群或者好友Id
// @param title string 标题
// @param digest string 摘要
// @param url string 链接
// @param retryCount int 重试次数
func SendPublicMsg(wxId, title, digest, url string, retryCount int) {
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SEND")); flag {
		return
	}

	if retryCount > 5 {
		log.Printf("重试五次失败，停止发送")
		return
	}

	// 组装参数
	param := map[string]any{
		"appName":  "FlyAI",
		"userName": "",
		"title":    title,
		"digest":   digest,
		"url":      url,
		"thumbUrl": "https://raw.githubusercontent.com/loks666/FlyAI/main/views/static/img/logo.png",
		"wxid":     wxId,
	}
	pbs, _ := json.Marshal(param)

	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		Post(config.Conf.Wechat.GetURL("/api/forwardPublicMsg"))
	if err != nil {
		log.Printf("发送公众号消息失败: %s", err.Error())
		// 休眠五秒后重新发送
		time.Sleep(5 * time.Second)
		SendPublicMsg(wxId, title, digest, url, retryCount+1)
	}
	log.Printf("发送公众号消息结果: %s", resp.String())
}

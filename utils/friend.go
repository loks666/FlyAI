package utils

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"go-wechat/common/constant"
	"go-wechat/config"
	"go-wechat/model/dto"
	"log"
	"slices"
	"strings"
)

// http客户端
var hc = resty.New()

// GetFriendList
// @description: 获取好友列表
// @return friends
// @return err
func GetFriendList() (friends []dto.FriendItem, err error) {
	var base dto.Response[[]dto.FriendItem]
	resp, err := hc.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetResult(&base).
		Post(config.Conf.Wechat.GetURL("/api/getContactList"))
	if err != nil {
		log.Printf("获取好友列表失败: %s", err.Error())
		return
	}
	log.Printf("获取好友列表结果: %s", resp.String())

	// 循环获取到的好友列表
	for _, item := range base.Data {
		// 跳过特殊账号
		// 跳过公众号和企业微信好友
		if strings.Contains(item.Wxid, "gh_") || strings.Contains(item.Wxid, "@openim") {
			continue
		}
		// 特殊Id跳过
		if slices.Contains(constant.SpecialId, item.Wxid) {
			continue
		}
		// 添加到待返回列表
		friends = append(friends, item)
	}

	return
}

// GetGroupMembers
// @description: 获取指定群成员
// @param wxId 群Id
// @return ids 群成员id数组
// @return members 群成员信息
// @return err 错误信息
func GetGroupMembers(wxId string) (ids []string, admin string, err error) {
	var base dto.Response[dto.GroupUser]
	// 组装参数
	param := map[string]any{
		"chatRoomId": wxId, // 群Id
	}
	pbs, _ := json.Marshal(param)

	_, err = hc.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		SetResult(&base).
		Post(config.Conf.Wechat.GetURL("/api/getMemberFromChatRoom"))
	if err != nil {
		log.Printf("获取群成员信息失败: %s", err.Error())
		return
	}
	// 昵称Id
	ids = strings.Split(base.Data.Members, "^G")
	admin = base.Data.Admin
	return
}

// GetContactProfile
// @description: 获取联系人信息
// @param wxId 微信Id
// @return info 联系人信息
// @return err 错误信息
func GetContactProfile(wxId string) (info dto.ContactProfile, err error) {
	var baseResp dto.Response[dto.ContactProfile]

	// 组装参数
	param := map[string]any{
		"wxid": wxId, // 群Id
	}
	pbs, _ := json.Marshal(param)

	_, err = hc.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		SetResult(&baseResp).
		Post(config.Conf.Wechat.GetURL("/api/getContactProfile"))
	if err != nil {
		log.Printf("获取成员详情失败: %s", err.Error())
		return
	}
	info = baseResp.Data
	return
}

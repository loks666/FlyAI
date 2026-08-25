package utils

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/loks666/FlyAI/common/constant"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/model/dto"
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
	if resp.IsError() {
		return nil, errors.New("微信桥接服务返回异常状态")
	}
	if base.Code != 1 {
		if base.Msg == "" {
			base.Msg = "微信桥接服务未返回有效通讯录"
		}
		return nil, errors.New(base.Msg)
	}
	if len(base.Data) == 0 {
		return nil, errors.New("微信桥接服务返回空通讯录，为避免误清理已保留现有数据")
	}

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
func GetGroupMembers(wxId string) (ids []string, nicknames map[string]string, admin string, err error) {
	var base dto.Response[dto.GroupUser]
	// 组装参数
	param := map[string]any{
		"chatRoomId": wxId, // 群Id
	}
	pbs, _ := json.Marshal(param)

	resp, err := hc.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(string(pbs)).
		SetResult(&base).
		Post(config.Conf.Wechat.GetURL("/api/getMemberFromChatRoom"))
	if err != nil {
		log.Printf("获取群成员信息失败: %s", err.Error())
		return
	}
	if resp.IsError() || base.Code != 1 {
		if base.Msg == "" {
			base.Msg = "微信桥接服务未返回有效群成员"
		}
		err = errors.New(base.Msg)
		return
	}
	// 群内昵称与成员Id按相同顺序返回；群内昵称为空时桥接层已回退个人昵称。
	for _, id := range strings.Split(base.Data.Members, "^G") {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		err = errors.New("当前群成员列表为空，为避免误清理已保留现有数据")
		return
	}
	nicknames = make(map[string]string, len(ids))
	memberNicknames := strings.Split(base.Data.MemberNickname, "^G")
	for index, id := range ids {
		if index < len(memberNicknames) && strings.TrimSpace(memberNicknames[index]) != "" {
			nicknames[id] = strings.TrimSpace(memberNicknames[index])
		}
	}
	admin = base.Data.Admin
	return
}

// GetGroupActiveMembers returns members who have sent any user message in the
// local WeChat history currently available on this machine.
func GetGroupActiveMembers(wxId string) (members []string, err error) {
	counts, err := GetGroupMessageCounts(wxId)
	if err != nil {
		return nil, err
	}
	for member := range counts {
		members = append(members, member)
	}
	return members, nil
}

// GetGroupMessageCounts returns every resolved member's user-message count
// from all group history currently readable by the local WeChat client.
func GetGroupMessageCounts(wxId string) (counts map[string]int64, err error) {
	var base dto.Response[dto.GroupActiveMembers]
	resp, err := hc.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetBody(map[string]any{"chatRoomId": wxId}).
		SetResult(&base).
		Post(config.Conf.Wechat.GetURL("/api/getChatRoomActiveMembers"))
	if err != nil {
		return nil, err
	}
	if resp.IsError() || base.Code != 1 {
		if base.Msg == "" {
			base.Msg = "微信桥接服务未返回历史发言成员"
		}
		return nil, errors.New(base.Msg)
	}
	if base.Data.Counts == nil {
		return nil, errors.New("微信桥接服务版本过低，未返回群成员消息条数")
	}
	return base.Data.Counts, nil
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

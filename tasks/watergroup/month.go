package watergroup

import (
	"fmt"
	"go-wechat/client"
	"go-wechat/config"
	"go-wechat/model/entity"
	"go-wechat/service"
	"go-wechat/utils"
	"log"
	"strings"
	"time"
)

// Month
// @description: 月排行榜
func Month() {
	groups, err := service.GetAllEnableChatRank()
	if err != nil {
		log.Printf("获取启用了聊天排行榜的群组失败, 错误信息: %v", err)
		return
	}

	for _, group := range groups {
		// 消息统计
		dealMonth(group.Wxid)

		res, ok := config.Conf.Resource["wordcloud"]
		if !ok {
			continue
		}

		// 获取上个月月份
		yd := time.Now().Local().AddDate(0, 0, -1).Format("200601")
		// 发送词云
		fileName := fmt.Sprintf("%s_%s.png", yd, group.Wxid)
		utils.SendImage(group.Wxid, fmt.Sprintf(res.Path, fileName), 0)
	}
}

// dealMonth
// @description: 处理请求
// @param gid
func dealMonth(gid string) {
	monthStr := time.Now().Local().AddDate(0, 0, -1).Format("2006年01月")
	notifyMsgs := []string{fmt.Sprintf("#%s水群排行榜", monthStr)}

	// 获取上月消息总数
	records, err := getRankData(gid, "month")
	if err != nil {
		log.Printf("获取上月消息排行失败, 错误信息: %v", err)
		return
	}
	log.Printf("上月消息总数: %+v", records)
	// 莫得消息，直接返回
	if len(records) == 0 {
		log.Printf("上月群[%s]无对话记录", gid)
		return
	}

	// 查询群成员总数
	var groupUsers int64
	err = client.MySQL.Model(&entity.GroupUser{}).
		Where("group_id = ?", gid).
		Where("is_member IS TRUE").
		Count(&groupUsers).Error
	if err != nil {
		log.Printf("查询群成员总数失败, 错误信息: %v", err)
	}
	// 计算活跃度
	showActivity := err == nil && groupUsers > 0
	activity := "0.00"
	if groupUsers > 0 {
		activity = fmt.Sprintf("%.2f", (float64(len(records))/float64(groupUsers))*100)
	}

	// 计算消息总数、中位数、前十位消息总数
	var msgCount, medianCount, topTenCount int64
	for idx, v := range records {
		msgCount += v.Count
		if idx == (len(records)/2)-1 {
			medianCount = v.Count
		}
		if len(records) > 10 && idx < 10 {
			topTenCount += v.Count
		}
	}
	// 计算活跃用户人均消息条数
	avgMsgCount := int(float64(msgCount) / float64(len(records)))

	// 组装消息总数推送信息
	notifyMsgs = append(notifyMsgs, " ")
	notifyMsgs = append(notifyMsgs, fmt.Sprintf("🗣️ %s本群 %d 位朋友共产生 %d 条发言", monthStr, len(records), msgCount))
	if showActivity {
		m := fmt.Sprintf("🎭 活跃度: %s%%，人均消息条数: %d，中位数: %d", activity, avgMsgCount, medianCount)
		// 计算前十占比
		if topTenCount > 0 {
			m += fmt.Sprintf("，前十名占比: %.2f%%", float64(topTenCount)/float64(msgCount)*100)
		}
		notifyMsgs = append(notifyMsgs, m)
	}
	notifyMsgs = append(notifyMsgs, "\n🏵 活跃用户排行榜 🏵")

	notifyMsgs = append(notifyMsgs, " ")
	for i, r := range records {
		// 只取前十条
		if i >= 10 {
			break
		}

		log.Printf("账号: %s[%s] -> %d", r.Nickname, r.GroupUser, r.Count)
		badge := "🏆"
		switch i {
		case 0:
			badge = "🥇"
		case 1:
			badge = "🥈"
		case 2:
			badge = "🥉"
		}
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条", badge, r.Nickname, r.Count))
	}

	notifyMsgs = append(notifyMsgs, fmt.Sprintf(" \n🎉感谢以上群友%s对群活跃做出的卓越贡献，也请未上榜的群友多多反思。", monthStr))

	log.Printf("排行榜: \n%s", strings.Join(notifyMsgs, "\n"))
	go utils.SendMessage(gid, "", strings.Join(notifyMsgs, "\n"), 0)
}

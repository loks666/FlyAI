package watergroup

import (
	"errors"
	"fmt"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
	"time"
)

var ErrNoWeekMessages = errors.New("最近7天没有可排行的消息")

// Week
// @description: 周排行榜
func Week() {
	groups, err := service.GetAllEnableChatRank()
	if err != nil {
		log.Printf("获取启用了聊天排行榜的群组失败, 错误信息: %v", err)
		return
	}

	for _, group := range groups {
		if _, err := SendWeekToGroup(group.Wxid); err != nil {
			log.Printf("发送群[%s]最近7天排行失败: %v", group.Wxid, err)
			continue
		}

		res, ok := config.Conf.Resource["wordcloud"]
		if !ok {
			continue
		}

		// 获取上周周数
		year, weekNo := time.Now().Local().AddDate(0, 0, -1).ISOWeek()
		// 发送词云
		fileName := fmt.Sprintf("%d%d_%s.png", year, weekNo, group.Wxid)
		utils.SendImage(group.Wxid, fmt.Sprintf(res.Path, fileName), 0)
	}
}

// BuildWeekRank builds the weekly ranking message without sending it.
func BuildWeekRank(gid string) (string, error) {
	notifyMsgs := []string{"#上周水群排行榜"}
	referenceTime := time.Now()
	weekStart, weekEnd := rollingWeekRange(referenceTime)
	// The local group-member table can contain historical leavers. Refresh it
	// from the signed-in WeChat client before calculating the denominator.
	service.SyncGroupMembers(gid)

	// 获取从生成时刻向前滚动7天的消息总数
	records, err := getRankData(gid, "week", referenceTime)
	if err != nil {
		return "", fmt.Errorf("获取最近7天消息排行失败: %w", err)
	}
	log.Printf("最近7天消息统计: %+v", records)
	// 莫得消息，直接返回
	if len(records) == 0 {
		return "", ErrNoWeekMessages
	}

	// 查询群成员总数
	var groupUsers int64
	err = client.MySQL.Model(&entity.GroupUser{}).
		Where("group_id = ?", gid).
		Where("is_member IS TRUE").
		Distinct("wxid").
		Count(&groupUsers).Error
	if err != nil {
		log.Printf("查询群成员总数失败, 错误信息: %v", err)
	}
	// 计算群成员参与情况
	showActivity := err == nil && groupUsers > 0
	activity := 0.0
	silentUsers := int64(0)
	if groupUsers > 0 {
		activity = float64(len(records)) / float64(groupUsers) * 100
		silentUsers = groupUsers - int64(len(records))
		if silentUsers < 0 {
			silentUsers = 0
		}
	}

	// 计算消息分布指标
	var msgCount, highFrequencyUsers int64
	var messageMix rankUser
	for _, v := range records {
		msgCount += v.Count
		messageMix.TextCount += v.TextCount
		messageMix.ImageCount += v.ImageCount
		messageMix.VoiceCount += v.VoiceCount
		messageMix.VideoCount += v.VideoCount
		messageMix.EmoticonCount += v.EmoticonCount
		messageMix.ShareCount += v.ShareCount
		messageMix.OtherCount += v.OtherCount
		if v.Count >= 20 {
			highFrequencyUsers++
		}
	}
	avgMsgCount := float64(msgCount) / float64(len(records))
	medianCount := float64(records[len(records)/2].Count)
	if len(records)%2 == 0 {
		medianCount = float64(records[len(records)/2-1].Count+records[len(records)/2].Count) / 2
	}
	maxUserCount := records[0].Count

	// 组装消息总数推送信息
	notifyMsgs = append(notifyMsgs, " ")
	notifyMsgs = append(notifyMsgs, fmt.Sprintf("📅 %s～%s",
		weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02")))
	notifyMsgs = append(notifyMsgs, fmt.Sprintf("🗣️ 最近7天本群 %d 位朋友共产生 %d 条发言", len(records), msgCount))
	if showActivity {
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("👥 群成员: %d 人，活跃: %d 人", groupUsers, len(records)))
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("💤 潜水: %d 人，活跃度: %.2f%%", silentUsers, activity))
	}
	notifyMsgs = append(notifyMsgs, fmt.Sprintf("📈 日均发言: %.1f 条，人均: %.1f 条，中位数: %.1f 条",
		float64(msgCount)/7, avgMsgCount, medianCount))
	notifyMsgs = append(notifyMsgs, "🧩 消息构成: "+formatMessageBreakdown(messageMix))
	notifyMsgs = append(notifyMsgs, fmt.Sprintf("🔥 高频发言者(≥20条): %d 人", highFrequencyUsers))
	notifyMsgs = append(notifyMsgs, fmt.Sprintf("👑 最高单人发言: %d 条", maxUserCount))
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
		notifyMsgs = append(notifyMsgs, fmt.Sprintf("%s %s -> %d条(%s)", badge, r.Nickname, r.Count, formatRankingBreakdown(r)))
	}

	notifyMsgs = append(notifyMsgs, " \n🎉感谢以上群友对群活跃做出的贡献🎉")
	notifyMsgs = append(notifyMsgs, "🤔 也请未上榜的群友多多反思🤔")

	message := strings.Join(notifyMsgs, "\n")
	log.Printf("排行榜: \n%s", message)
	return message, nil
}

func formatMessageBreakdown(record rankUser) string {
	parts := make([]string, 0, 7)
	items := []struct {
		label string
		count int64
	}{
		{"文本", record.TextCount},
		{"图片", record.ImageCount},
		{"视频", record.VideoCount},
		{"表情", record.EmoticonCount},
		{"转发", record.ShareCount},
		{"语音", record.VoiceCount},
		{"其他", record.OtherCount},
	}
	for _, item := range items {
		if item.count > 0 {
			parts = append(parts, fmt.Sprintf("%s%d", item.label, item.count))
		}
	}
	if len(parts) == 0 {
		return "暂无分类"
	}
	return strings.Join(parts, ", ")
}

func formatRankingBreakdown(record rankUser) string {
	parts := make([]string, 0, 7)
	items := []struct {
		emoji string
		count int64
	}{
		{"📝", record.TextCount},
		{"🖼️", record.ImageCount},
		{"🎬", record.VideoCount},
		{"😄", record.EmoticonCount},
		{"🔁", record.ShareCount},
		{"🎙️", record.VoiceCount},
		{"📦", record.OtherCount},
	}
	for _, item := range items {
		if item.count > 0 {
			parts = append(parts, fmt.Sprintf("%s%d", item.emoji, item.count))
		}
	}
	if len(parts) == 0 {
		return "📦0"
	}
	return strings.Join(parts, " , ")
}

// SendWeekToGroup builds and synchronously sends one group's weekly ranking.
func SendWeekToGroup(gid string) (string, error) {
	message, err := BuildWeekRank(gid)
	if err != nil {
		return "", err
	}
	if err := utils.SendMessage(gid, "", message, 0); err != nil {
		return "", err
	}
	return message, nil
}

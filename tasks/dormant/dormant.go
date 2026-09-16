package dormant

import (
	"fmt"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultDormantDisplayLimit = 20
	compactDormantDisplayLimit = 10
	maxDormantReportRunes      = 2000
)

type reportData struct {
	GeneratedAt  time.Time
	CurrentCount int64
	Threshold    int
	Members      []service.DormantMember
}

func BuildReport(groupId string) (string, error) {
	report, err := buildReportData(groupId)
	if err != nil {
		return "", err
	}
	return formatReport(report.GeneratedAt, report.CurrentCount, report.Threshold, report.Members), nil
}

func buildReportData(groupId string) (reportData, error) {
	if err := service.SyncGroupMembers(groupId); err != nil {
		return reportData{}, fmt.Errorf("同步当前群成员失败: %w", err)
	}
	messageCounts, err := utils.GetGroupMessageCounts(groupId)
	if err != nil {
		return reportData{}, fmt.Errorf("读取群历史消息条数失败: %w", err)
	}
	setting, err := service.GetFeatureSetting(service.FeatureDormant)
	if err != nil {
		return reportData{}, fmt.Errorf("读取潜水消息阈值失败: %w", err)
	}
	members, err := service.GetDormantGroupMembers(groupId, messageCounts, current.GetRobotInfo().WxId, setting.MessageThreshold)
	if err != nil {
		return reportData{}, fmt.Errorf("统计潜水成员失败: %w", err)
	}
	currentCount, err := service.CountCurrentGroupMembers(groupId)
	if err != nil {
		return reportData{}, fmt.Errorf("统计当前群成员失败: %w", err)
	}
	return reportData{
		GeneratedAt:  time.Now(),
		CurrentCount: currentCount,
		Threshold:    setting.MessageThreshold,
		Members:      members,
	}, nil
}

func formatReport(now time.Time, currentCount int64, threshold int, members []service.DormantMember) string {
	message := formatReportWithLimit(now, currentCount, threshold, members, defaultDormantDisplayLimit)
	if utf8.RuneCountInString(message) > maxDormantReportRunes {
		message = formatReportWithLimit(now, currentCount, threshold, members, compactDormantDisplayLimit)
	}
	return message
}

func formatReportWithLimit(now time.Time, currentCount int64, threshold int, members []service.DormantMember, limit int) string {
	lines := []string{
		"#潜水成员统计",
		"",
		fmt.Sprintf("📅 截至 %s", now.Format("2006-01-02")),
		fmt.Sprintf("👥 当前群成员: %d 人", currentCount),
		fmt.Sprintf("💤 本机可读取历史中发言少于%d条: %d 人", threshold, len(members)),
	}
	if len(members) == 0 {
		return strings.Join(append(lines, "", fmt.Sprintf("🎉 当前没有发言少于%d条的成员", threshold)), "\n")
	}
	if limit > len(members) {
		limit = len(members)
	}
	lines = append(lines, "", fmt.Sprintf("📋 仅展示前%d人昵称：", limit))
	for index, member := range members[:limit] {
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, dormantDisplayName(member)))
	}
	return strings.Join(lines, "\n")
}

func dormantDisplayName(member service.DormantMember) string {
	for _, name := range []string{
		member.GroupNickname,
		member.WechatNickname,
		member.Nickname,
	} {
		name = strings.TrimSpace(name)
		if name != "" {
			return strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(name)
		}
	}
	return "（昵称未获取）"
}

func SendReportToGroup(groupId string) (string, error) {
	report, err := buildReportData(groupId)
	if err != nil {
		return "", err
	}
	message := formatReport(report.GeneratedAt, report.CurrentCount, report.Threshold, report.Members)
	if err = utils.SendMessage(groupId, "", message, 0); err != nil {
		return message, err
	}
	return message, nil
}

func Weekly() {
	groups, err := service.GetAllEnableDormant()
	if err != nil {
		log.Printf("获取启用潜水统计的群失败: %v", err)
		return
	}
	for _, group := range groups {
		if _, err = SendReportToGroup(group.Wxid); err != nil {
			log.Printf("发送群[%s]潜水成员统计失败: %v", group.Wxid, err)
		}
		time.Sleep(time.Second)
	}
}

package dormant

import (
	"fmt"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// WeChat does not publish a stable Windows-client text limit. Keep enough
// headroom below the commonly observed boundary and fall back to a TXT file.
const maxInlineReportRunes = 1800

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
	lines := []string{
		"#潜水成员统计",
		"",
		fmt.Sprintf("📅 截至 %s", now.Format("2006-01-02")),
		fmt.Sprintf("👥 当前群成员: %d 人", currentCount),
		fmt.Sprintf("💤 从入群起发言少于%d条: %d 人", threshold, len(members)),
	}
	if len(members) == 0 {
		return strings.Join(append(lines, "", fmt.Sprintf("🎉 当前没有发言少于%d条的成员", threshold)), "\n")
	}
	names := make([]string, 0, len(members))
	for _, member := range members {
		name := strings.TrimSpace(member.Nickname)
		if name == "" {
			name = member.Wxid
		}
		names = append(names, fmt.Sprintf("%s（%d条）", name, member.Count))
	}
	return strings.Join(append(lines, "", strings.Join(names, "、")), "\n")
}

func formatSummary(now time.Time, currentCount int64, threshold int, memberCount int) string {
	return strings.Join([]string{
		"#潜水成员统计",
		"",
		fmt.Sprintf("📅 截至 %s", now.Format("2006-01-02")),
		fmt.Sprintf("👥 当前群成员: %d 人", currentCount),
		fmt.Sprintf("💤 从入群起发言少于%d条: %d 人", threshold, memberCount),
		"",
		"📎 名单较长，完整成员资料请查看 TXT 文件",
	}, "\n")
}

func requiresTextAttachment(message string) bool {
	return utf8.RuneCountInString(message) > maxInlineReportRunes
}

func formatField(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "（未获取）"
	}
	return strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(value)
}

func formatTextReport(now time.Time, currentCount int64, threshold int, members []service.DormantMember) string {
	lines := []string{
		"FlyAI 潜水成员名单",
		fmt.Sprintf("生成时间：%s", now.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("当前群成员：%d 人", currentCount),
		fmt.Sprintf("统计条件：从入群起发言少于 %d 条", threshold),
		fmt.Sprintf("潜水成员：%d 人", len(members)),
		"",
		"序号\t群昵称\t微信昵称\t微信号\twxid\t发言数",
	}
	for index, member := range members {
		lines = append(lines, fmt.Sprintf(
			"%d\t%s\t%s\t%s\t%s\t%d",
			index+1,
			formatField(member.GroupNickname),
			formatField(member.WechatNickname),
			formatField(member.WechatAccount),
			formatField(member.Wxid),
			member.Count,
		))
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}

func writeTextReport(report reportData) (string, error) {
	directory := filepath.Join(os.TempDir(), "FlyAI", "dormant-reports")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("创建潜水报告临时目录失败: %w", err)
	}
	file, err := os.CreateTemp(directory, fmt.Sprintf("FlyAI-潜水名单-%s-*.txt", report.GeneratedAt.Format("20060102-150405")))
	if err != nil {
		return "", fmt.Errorf("创建潜水报告临时文件失败: %w", err)
	}
	path := file.Name()
	content := []byte("\xEF\xBB\xBF" + formatTextReport(report.GeneratedAt, report.CurrentCount, report.Threshold, report.Members))
	if _, err = file.Write(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("写入潜水报告临时文件失败: %w", err)
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("关闭潜水报告临时文件失败: %w", err)
	}
	return path, nil
}

func SendReportToGroup(groupId string) (string, error) {
	report, err := buildReportData(groupId)
	if err != nil {
		return "", err
	}
	message := formatReport(report.GeneratedAt, report.CurrentCount, report.Threshold, report.Members)
	var inlineErr error
	if !requiresTextAttachment(message) {
		inlineErr = utils.SendMessage(groupId, "", message, 0)
		if inlineErr == nil {
			return message, nil
		}
		log.Printf("群[%s]潜水名单文本发送失败，改用 TXT: %v", groupId, inlineErr)
	}

	path, err := writeTextReport(report)
	if err != nil {
		if inlineErr != nil {
			return message, fmt.Errorf("文本发送失败（%v），且生成 TXT 失败: %w", inlineErr, err)
		}
		return message, err
	}
	defer os.Remove(path)

	summary := formatSummary(report.GeneratedAt, report.CurrentCount, report.Threshold, len(report.Members))
	summaryErr := utils.SendMessage(groupId, "", summary, 0)
	if err = utils.SendFile(groupId, path, 0); err != nil {
		if summaryErr != nil {
			return message, fmt.Errorf("摘要发送失败（%v），TXT 发送也失败: %w", summaryErr, err)
		}
		return message, fmt.Errorf("摘要已发送，但潜水名单 TXT 发送失败: %w", err)
	}
	return summary, nil
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

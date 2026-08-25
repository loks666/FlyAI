package dormant

import (
	"fmt"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
	"time"
)

func BuildReport(groupId string) (string, error) {
	if err := service.SyncGroupMembers(groupId); err != nil {
		return "", fmt.Errorf("同步当前群成员失败: %w", err)
	}
	messageCounts, err := utils.GetGroupMessageCounts(groupId)
	if err != nil {
		return "", fmt.Errorf("读取群历史消息条数失败: %w", err)
	}
	setting, err := service.GetFeatureSetting(service.FeatureDormant)
	if err != nil {
		return "", fmt.Errorf("读取潜水消息阈值失败: %w", err)
	}
	members, err := service.GetDormantGroupMembers(groupId, messageCounts, current.GetRobotInfo().WxId, setting.MessageThreshold)
	if err != nil {
		return "", fmt.Errorf("统计潜水成员失败: %w", err)
	}
	currentCount, err := service.CountCurrentGroupMembers(groupId)
	if err != nil {
		return "", fmt.Errorf("统计当前群成员失败: %w", err)
	}
	return formatReport(time.Now(), currentCount, setting.MessageThreshold, members), nil
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

func SendReportToGroup(groupId string) (string, error) {
	message, err := BuildReport(groupId)
	if err != nil {
		return "", err
	}
	if err = utils.SendMessage(groupId, "", message, 0); err != nil {
		return "", err
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

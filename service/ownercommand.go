package service

import "strings"

const (
	OwnerCommandWeekly  = "weekly"
	OwnerCommandDormant = "dormant"
	OwnerCommandNews    = "news"
)

// ParseOwnerCommand accepts only an explicit @robot-name prefix followed by
// one of the fixed owner commands. Unicode spaces inserted by WeChat are
// handled by strings.TrimSpace.
func ParseOwnerCommand(content, robotName string) (string, bool) {
	content = strings.TrimSpace(content)
	robotName = strings.TrimSpace(robotName)
	if content == "" || robotName == "" {
		return "", false
	}
	prefix := "@" + robotName
	if !strings.HasPrefix(content, prefix) {
		return "", false
	}
	commandText := strings.TrimSpace(strings.TrimPrefix(content, prefix))
	switch commandText {
	case "/周报":
		return OwnerCommandWeekly, true
	case "/潜水":
		return OwnerCommandDormant, true
	case "/热点":
		return OwnerCommandNews, true
	default:
		return "", false
	}
}

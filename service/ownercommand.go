package service

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	OwnerCommandYesterday = "yesterday"
	OwnerCommandWeekly    = "weekly"
	OwnerCommandMonthly   = "monthly"
	OwnerCommandYearly    = "yearly"
	OwnerCommandDormant   = "dormant"
	OwnerCommandNews      = "news"
	OwnerCommandAttack    = "attack"
)

type OwnerAttackMode string

const (
	OwnerAttackChinese OwnerAttackMode = "chinese"
	OwnerAttackMorse   OwnerAttackMode = "morse"
)

// ParseOwnerCommand accepts one of the fixed owner commands found after an
// explicit @robot-name prefix. Unicode spaces inserted by WeChat are handled
// by strings.TrimSpace.
func ParseOwnerCommand(content, robotName string) (string, bool) {
	commandText, ok := ExtractRobotCommand(content, robotName)
	if !ok {
		return "", false
	}
	switch commandText {
	case "/昨排":
		return OwnerCommandYesterday, true
	case "/排行榜", "/周排":
		return OwnerCommandWeekly, true
	case "/月排":
		return OwnerCommandMonthly, true
	case "/年排":
		return OwnerCommandYearly, true
	case "/潜水":
		return OwnerCommandDormant, true
	case "/热点":
		return OwnerCommandNews, true
	default:
		return "", false
	}
}

// ParseOwnerAttackTargets accepts an optional robot mention followed by
// 给我骂他 and one or more @names. WeChat Unicode spaces, commas, and adjacent
// mentions are all supported.
func ParseOwnerAttackTargets(content, robotName string) ([]string, OwnerAttackMode, bool) {
	remainder := strings.TrimSpace(content)
	robotName = strings.TrimSpace(robotName)
	if robotName != "" {
		for _, prefix := range []string{"@@" + robotName, "@" + robotName} {
			if strings.HasPrefix(remainder, prefix) {
				remainder = strings.TrimSpace(strings.TrimPrefix(remainder, prefix))
				break
			}
		}
	}
	const instruction = "给我骂他"
	if !strings.HasPrefix(remainder, instruction) {
		return nil, "", false
	}
	targetText := strings.TrimSpace(strings.TrimPrefix(remainder, instruction))
	mode := OwnerAttackChinese
	if strings.HasPrefix(targetText, ".") {
		mode = OwnerAttackMorse
		targetText = strings.TrimSpace(strings.TrimPrefix(targetText, "."))
	}
	if !strings.HasPrefix(targetText, "@") {
		return nil, "", false
	}
	parts := strings.Split(targetText, "@")
	targets := make([]string, 0, len(parts)-1)
	seen := make(map[string]struct{})
	for _, part := range parts[1:] {
		name := strings.Trim(part, " \t\r\n\u2005,，、")
		if name == "" {
			return nil, "", false
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		targets = append(targets, name)
	}
	return targets, mode, len(targets) > 0
}

// ExtractRobotCommand separates mention recognition from command matching.
// Once the message starts with @robot-name, the first slash-prefixed token in
// the remaining text is treated as a command candidate.
func ExtractRobotCommand(content, robotName string) (string, bool) {
	remainder, ok := robotMentionRemainder(content, robotName)
	if !ok {
		return "", false
	}
	slashIndex := strings.IndexRune(remainder, '/')
	if slashIndex < 0 {
		return "", false
	}
	fields := strings.Fields(strings.TrimSpace(remainder[slashIndex:]))
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "/") {
		return "", false
	}
	return fields[0], true
}

func HasRobotCommand(content, robotName string) bool {
	_, ok := ExtractRobotCommand(content, robotName)
	return ok
}

// IsRobotMention reports whether the message starts with an explicit @robot
// mention. A whitespace boundary prevents names such as "Fly AI2" from being
// treated as a mention of "Fly AI".
func IsRobotMention(content, robotName string) bool {
	_, ok := robotMentionRemainder(content, robotName)
	return ok
}

func robotMentionRemainder(content, robotName string) (string, bool) {
	content = strings.TrimSpace(content)
	robotName = strings.TrimSpace(robotName)
	if content == "" || robotName == "" {
		return "", false
	}
	prefix := "@" + robotName
	if !strings.HasPrefix(content, prefix) {
		return "", false
	}
	remainder := strings.TrimPrefix(content, prefix)
	if remainder == "" {
		return "", true
	}
	first, _ := utf8.DecodeRuneInString(remainder)
	if !unicode.IsSpace(first) {
		return "", false
	}
	return strings.TrimSpace(remainder), true
}

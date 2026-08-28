package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/loks666/FlyAI/plugin"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

var attackRepliesPath = filepath.Join("utils", "attack.json")

// AttackReply applies one global state machine per sender wxid, regardless of
// which group the sender uses to mention the robot.
func AttackReply(ctx *plugin.MessageContext) {
	ctx.Abort()
	action := service.BeginNonOwnerReply(ctx.GroupUser)
	if action == service.NonOwnerReplyIgnore {
		return
	}
	reply, err := nonOwnerMentionReply(action, attackRepliesPath)
	if err != nil {
		log.Printf("读取非主人回复语料失败: %v", err)
		service.CompleteNonOwnerReply(ctx.GroupUser, action, false)
		return
	}
	err = utils.SendMessage(ctx.FromUser, ctx.GroupUser, reply, 0)
	if err == nil && action == service.NonOwnerReplyBalanceAndMute {
		if emotionErr := utils.SendFavoriteEmotion(ctx.FromUser); emotionErr != nil {
			log.Printf("发送余额不足后的收藏表情失败: %v", emotionErr)
		}
	}
	service.CompleteNonOwnerReply(ctx.GroupUser, action, err == nil)
	if err != nil {
		log.Printf("发送非管理员回复失败: %v", err)
	}
}

func nonOwnerMentionReply(action service.NonOwnerReplyAction, path string) (string, error) {
	if action == service.NonOwnerReplyBalanceAndMute {
		return administratorRechargeReply(), nil
	}
	return randomAttackReply(path)
}

func administratorRechargeReply() string {
	owners, err := service.GetOwners()
	if err != nil {
		log.Printf("读取管理员资料失败: %v", err)
	}
	nickname := "尚未绑定"
	wechatAccount := "尚未配置"
	if len(owners) > 0 {
		nickname = firstNonBlank(owners[0].Nickname, owners[0].Wxid)
		wechatAccount = firstNonBlank(owners[0].WechatAccount, owners[0].Wxid)
	}
	return formatAdministratorRechargeReply(nickname, wechatAccount)
}

func formatAdministratorRechargeReply(nickname, wechatAccount string) string {
	return fmt.Sprintf("您的余额不足，请联系管理员【%s，%s】充值，未来一小时内不再回复您的消息", nickname, wechatAccount)
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func randomAttackReply(path string) (string, error) {
	replies, err := loadAttackReplies(path, service.OwnerAttackChinese)
	if err != nil {
		return "", err
	}
	return replies[rand.Intn(len(replies))], nil
}

type attackReplyLists struct {
	Chinese []string `json:"chinese"`
	Morse   []string `json:"morse"`
}

func loadAttackReplies(path string, mode service.OwnerAttackMode) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lists attackReplyLists
	if err = json.Unmarshal(data, &lists); err != nil {
		// Keep old single-array files readable during upgrades.
		var legacy []string
		if legacyErr := json.Unmarshal(data, &legacy); legacyErr != nil {
			return nil, fmt.Errorf("解析 JSON 失败: %w", err)
		}
		lists.Chinese = legacy
	}
	replies := lists.Chinese
	if mode == service.OwnerAttackMorse {
		replies = lists.Morse
	}
	valid := replies[:0]
	for _, reply := range replies {
		if reply = strings.TrimSpace(reply); reply != "" {
			valid = append(valid, reply)
		}
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("%s 列表中没有可用文本", mode)
	}
	return valid, nil
}

func sendRandomAttackReplies(toId string, targetNames []string, mode service.OwnerAttackMode) (string, error) {
	replies, err := loadAttackReplies(attackRepliesPath, mode)
	if err != nil {
		return "", err
	}
	lines := make([]string, 0, len(targetNames))
	for _, targetName := range targetNames {
		lines = append(lines, formatTargetAttackReply(targetName, replies[rand.Intn(len(replies))]))
	}
	message := strings.Join(lines, "\n")
	return message, utils.SendMessage(toId, "", message, 0)
}

func formatTargetAttackReply(targetName, reply string) string {
	return fmt.Sprintf("@%s %s", strings.TrimSpace(targetName), strings.TrimSpace(reply))
}

package plugins

import (
	"fmt"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/plugin"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/tasks/dormant"
	"github.com/loks666/FlyAI/tasks/news"
	"github.com/loks666/FlyAI/tasks/watergroup"
	"github.com/loks666/FlyAI/utils"
	"sync"
)

var ownerCommandRunning sync.Map

var (
	sendYesterdayByOwner = watergroup.SendYesterdayToGroup
	sendWeeklyByOwner    = watergroup.SendWeekToGroup
	sendMonthlyByOwner   = watergroup.SendMonthToGroup
	sendYearlyByOwner    = watergroup.SendYearToGroup
	sendDormantByOwner   = dormant.SendReportToGroup
	sendNewsByOwner      = news.SendMorningPostToGroup
	sendAttackByOwner    = sendRandomAttackReplies
)

// OwnerCommand runs the three fixed group functions for a bound owner.
func OwnerCommand(ctx *plugin.MessageContext) {
	ctx.Abort()
	if !service.IsOwnerWxid(ctx.GroupUser) {
		return
	}
	robotName := current.GetRobotInfo().Name
	command := ""
	targetNames, attackMode, attackCommand := service.ParseOwnerAttackTargets(ctx.Content, robotName)
	if attackCommand {
		command = service.OwnerCommandAttack
	} else {
		commandText, hasCommand := service.ExtractRobotCommand(ctx.Content, robotName)
		if !hasCommand {
			return
		}
		var ok bool
		command, ok = service.ParseOwnerCommand(ctx.Content, robotName)
		if !ok {
			_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser,
				fmt.Sprintf("未识别命令：%s\n可用命令：/昨排、/周排、/月排、/年排、/排行榜、/潜水、/热点、给我骂他@目标昵称", commandText), 0)
			return
		}
	}

	key := ctx.FromUser + ":" + command
	if _, loaded := ownerCommandRunning.LoadOrStore(key, true); loaded {
		_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser, "该任务正在执行，请稍候。", 0)
		return
	}
	defer ownerCommandRunning.Delete(key)

	group, err := service.GetFriendInfoById(ctx.FromUser)
	if err != nil || !group.IsOk {
		_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser, "未找到当前群聊，请先同步通讯录。", 0)
		return
	}

	err = runOwnerCommand(group.Wxid, targetNames, attackMode, command)
	if err != nil {
		_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser, "执行失败："+err.Error(), 0)
	}
}

func runOwnerCommand(groupId string, targetNames []string, attackMode service.OwnerAttackMode, command string) error {
	switch command {
	case service.OwnerCommandYesterday:
		_, err := sendYesterdayByOwner(groupId)
		return err
	case service.OwnerCommandWeekly:
		_, err := sendWeeklyByOwner(groupId)
		return err
	case service.OwnerCommandMonthly:
		_, err := sendMonthlyByOwner(groupId)
		return err
	case service.OwnerCommandYearly:
		_, err := sendYearlyByOwner(groupId)
		return err
	case service.OwnerCommandDormant:
		_, err := sendDormantByOwner(groupId)
		return err
	case service.OwnerCommandNews:
		_, err := sendNewsByOwner(groupId)
		return err
	case service.OwnerCommandAttack:
		_, err := sendAttackByOwner(groupId, targetNames, attackMode)
		return err
	default:
		return fmt.Errorf("不支持的命令")
	}
}

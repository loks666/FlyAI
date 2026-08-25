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
	sendWeeklyByOwner  = watergroup.SendWeekToGroup
	sendDormantByOwner = dormant.SendReportToGroup
	sendNewsByOwner    = news.SendMorningPostToGroup
)

// OwnerCommand runs the three fixed group functions for a bound owner.
func OwnerCommand(ctx *plugin.MessageContext) {
	ctx.Abort()
	command, ok := service.ParseOwnerCommand(ctx.Content, current.GetRobotInfo().Name)
	if !ok || !service.IsOwnerWxid(ctx.GroupUser) {
		return
	}

	key := ctx.FromUser + ":" + command
	if _, loaded := ownerCommandRunning.LoadOrStore(key, true); loaded {
		_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser, "该任务正在执行，请稍候。", 0)
		return
	}
	defer ownerCommandRunning.Delete(key)

	group, err := service.GetFriendInfoById(ctx.FromUser)
	if err != nil || !group.IsOk || !group.EnableMonitor {
		_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser, "该群未开启监控。", 0)
		return
	}

	err = runOwnerCommand(group.Wxid, command)
	if err != nil {
		_ = utils.SendMessage(ctx.FromUser, ctx.GroupUser, "执行失败："+err.Error(), 0)
	}
}

func runOwnerCommand(groupId, command string) error {
	switch command {
	case service.OwnerCommandWeekly:
		_, err := sendWeeklyByOwner(groupId)
		return err
	case service.OwnerCommandDormant:
		_, err := sendDormantByOwner(groupId)
		return err
	case service.OwnerCommandNews:
		_, err := sendNewsByOwner(groupId)
		return err
	default:
		return fmt.Errorf("不支持的命令")
	}
}

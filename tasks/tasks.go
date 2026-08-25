package tasks

import (
	"github.com/go-co-op/gocron"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/tasks/cleargroupuser"
	"github.com/loks666/FlyAI/tasks/friends"
	"github.com/loks666/FlyAI/tasks/hottop"
	"github.com/loks666/FlyAI/tasks/summary"
	"github.com/loks666/FlyAI/tasks/watergroup"
	"log"
	"time"
)

// InitTasks
// @description: 初始化定时任务
func InitTasks() {
	// 没启用直接返回
	if !config.Conf.Task.Enable {
		log.Println("未启用定时任务")
		return
	}
	// 定时任务发送消息
	s := gocron.NewScheduler(time.Local)

	// 水群排行
	if config.Conf.Task.WaterGroup.Enable {
		log.Printf("水群排行任务已启用，执行表达式: %+v", config.Conf.Task.WaterGroup.Cron)
		if config.Conf.Task.WaterGroup.Cron.Yesterday != "" {
			_, _ = s.Cron(config.Conf.Task.WaterGroup.Cron.Yesterday).Do(watergroup.Yesterday)
		}
		if config.Conf.Task.WaterGroup.Cron.Month != "" {
			_, _ = s.Cron(config.Conf.Task.WaterGroup.Cron.Month).Do(watergroup.Month)
		}
		if config.Conf.Task.WaterGroup.Cron.Year != "" {
			_, _ = s.Cron(config.Conf.Task.WaterGroup.Cron.Year).Do(watergroup.Year)
		}
	}

	// 群聊总结
	if config.Conf.Task.GroupSummary.Enable {
		log.Printf("群聊总结任务已启用，执行表达式: %s", config.Conf.Task.GroupSummary.Cron)
		_, _ = s.Cron(config.Conf.Task.GroupSummary.Cron).Do(summary.AiSummary)
	}

	// 更新好友列表
	if config.Conf.Task.SyncFriends.Enable {
		log.Printf("更新好友列表任务已启用，执行表达式: %s", config.Conf.Task.SyncFriends.Cron)
		_, _ = s.Cron(config.Conf.Task.SyncFriends.Cron).Do(service.SyncFriendCatalog)
		go service.SyncFriendCatalog()
	}

	// 可扩展定时功能由数据库中的统一时间驱动，修改后无需重启。
	_, _ = s.Cron("* * * * *").Do(RunTimedFeatures)

	// 热榜
	if config.Conf.Task.HotTop.Enable {
		_, _ = s.Cron(config.Conf.Task.HotTop.Cron).Do(hottop.HotTop)
	}

	// 每天0点检查一次处理清理群成员
	_, _ = s.Cron("0 0 * * *").Do(cleargroupuser.ClearGroupUser)
	// 每天0点清空AI日使用次数
	_, _ = s.Cron("0 0 * * *").Do(friends.ClearAiUsedToday)

	// 开启定时任务
	s.StartAsync()
	log.Println("定时任务初始化成功")
}

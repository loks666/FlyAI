package tasks

import (
	"github.com/go-co-op/gocron"
	"go-wechat/config"
	"go-wechat/tasks/cleargroupuser"
	"go-wechat/tasks/friends"
	"go-wechat/tasks/hottop"
	"go-wechat/tasks/news"
	"go-wechat/tasks/summary"
	"go-wechat/tasks/watergroup"
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
		if config.Conf.Task.WaterGroup.Cron.Week != "" {
			_, _ = s.Cron(config.Conf.Task.WaterGroup.Cron.Week).Do(watergroup.Week)
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
		_, _ = s.Cron(config.Conf.Task.SyncFriends.Cron).Do(friends.Sync)
	}

	// 每日早报
	if config.Conf.Task.News.Enable {
		_, _ = s.Cron(config.Conf.Task.News.Cron).Do(news.DailyNews)
	}

	// 热榜
	if config.Conf.Task.HotTop.Enable {
		_, _ = s.Cron(config.Conf.Task.HotTop.Cron).Do(hottop.HotTop)
	}

	// 每天0点检查一次处理清理群成员
	_, _ = s.Cron("0 0 * * *").Do(cleargroupuser.ClearGroupUser)

	// 开启定时任务
	s.StartAsync()
	log.Println("定时任务初始化成功")
}

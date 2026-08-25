package hottop

import (
	"fmt"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/common/types"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"slices"
	"strings"
	"time"
)

// HotTop
// @description: 热搜排行榜
func HotTop() {
	// 发送到开启了的群
	groups, err := service.GetAllEnableHotTop()
	if err != nil {
		log.Printf("获取启用了热榜的群组失败, 错误信息: %v", err)
		return
	}
	// 获取热榜数据
	news := getTopData()
	if len(news) == 0 {
		return
	}

	// 组装消息
	msg := fmt.Sprintf("#热搜排行榜\n \n嘿，朋友，有新的新闻了喔，快来康康吧\n \n%s", strings.Join(news, "\n-------\n"))
	for _, group := range groups {
		utils.SendMessage(group.Wxid, "", msg, 0)
		// 休眠一秒，防止频繁发送
		time.Sleep(time.Second)
	}
}

// getTopData
// @description: 获取热榜数据
// @return data
func getTopData() (data []string) {
	// 获取热榜数据
	records := utils.NewsUtil().GetHotTop()
	if len(records) == 0 {
		log.Println("获取热榜数据失败")
		return
	}
	var datas = make([]entity.HotTop, 0)
	for _, item := range records {
		var d = entity.HotTop{
			CreatedAt: types.DateTime(time.Now().Local()),
			Title:     item.Title,
			Hot:       item.Hot,
			Url:       item.Url,
			MobileUrl: item.MobilUrl,
		}
		datas = append(datas, d)
	}

	// 获取缓存数据
	var oldTitles []string
	err := client.MySQL.Model(&entity.HotTop{}).Order("created_at DESC").Limit(len(datas)).Pluck("title", &oldTitles).Error
	if err != nil {
		log.Println("获取历史热榜数据失败", err)
		return
	}

	// 筛选出新数据
	var newDatas []entity.HotTop
	for _, d := range datas {
		if slices.Contains(oldTitles, d.Title) {
			continue
		}
		d.Channel = "百度"
		newDatas = append(newDatas, d)
		data = append(data, fmt.Sprintf("标题: %s\n热度: %s\n详情: %s", d.Title, d.Hot, d.Url))
	}
	// 保存新数据到数据库
	if len(newDatas) > 0 {
		err = client.MySQL.Create(&newDatas).Error
		if err != nil {
			log.Println("保存新热榜数据失败", err)
		}
	}
	return
}

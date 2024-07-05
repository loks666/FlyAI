package news

import (
	"fmt"
	"go-wechat/service"
	"go-wechat/utils"
	"log"
	"strings"
	"time"
)

// DailyNews
// @description: 每日新闻
func DailyNews() {
	groups, err := service.GetAllEnableNews()
	if err != nil {
		log.Printf("获取启用了聊天排行榜的群组失败, 错误信息: %v", err)
		return
	}

	news := utils.NewsUtil().MorningPost()
	if len(news) == 0 {
		log.Println("每日早报获取失败")
		return
	}

	newsStr := fmt.Sprintf("#每日早报\n\n又是新的一天了，让我们康康一觉醒来世界又发生了哪些变化~\n\n%s", strings.Join(news, "\n \n"))

	// 循环发送新闻
	for _, group := range groups {
		// 发送消息
		utils.SendMessage(group.Wxid, "", newsStr, 0)
		// 休眠一秒，防止频繁发送
		time.Sleep(time.Second)
	}
}

package news

import (
	"errors"
	"fmt"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"log"
	"strings"
	"time"
)

var ErrNoHotNews = errors.New("暂未获取到实时热点")

// BuildMorningPost creates the Chinese hot-list morning brief without sending it.
func BuildMorningPost() (string, error) {
	items := utils.NewsUtil().MorningPost()
	if len(items) == 0 {
		return "", ErrNoHotNews
	}
	return formatMorningPost(items, time.Now()), nil
}

func formatMorningPost(items []string, now time.Time) string {
	lines := make([]string, 0, len(items))
	for index, item := range items {
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, item))
	}
	return fmt.Sprintf("#📅%s日实时热点早报\n\n%s", now.Format("2006-01-02"), strings.Join(lines, "\n"))
}

// SendMorningPostToGroup builds and sends one group's hot-list morning brief.
func SendMorningPostToGroup(groupId string) (string, error) {
	message, err := BuildMorningPost()
	if err != nil {
		return "", err
	}
	if err = utils.SendMessage(groupId, "", message, 0); err != nil {
		return "", err
	}
	return message, nil
}

// DailyNews
// @description: 每日新闻
func DailyNews() {
	groups, err := service.GetAllEnableNews()
	if err != nil {
		log.Printf("获取启用了每日热点早报的群组失败, 错误信息: %v", err)
		return
	}

	newsStr, err := BuildMorningPost()
	if err != nil {
		log.Printf("每日热点早报获取失败: %v", err)
		return
	}

	// 循环发送新闻
	for _, group := range groups {
		// 发送消息
		if err = utils.SendMessage(group.Wxid, "", newsStr, 0); err != nil {
			log.Printf("发送群[%s]每日热点早报失败: %v", group.Wxid, err)
		}
		// 休眠一秒，防止频繁发送
		time.Sleep(time.Second)
	}
}

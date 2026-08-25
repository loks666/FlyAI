package utils

import (
	"github.com/go-resty/resty/v2"
	"github.com/loks666/FlyAI/model/dto"
	"html"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

// News provides the daily Chinese hot-list brief and the legacy hot-list surface.
type News interface {
	MorningPost() []string
	GetHotTop() []dto.HotTopDataItem
}

type news struct{}

func NewsUtil() News {
	return &news{}
}

type hotboardItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type hotboardResponse struct {
	Type       string         `json:"type"`
	UpdateTime string         `json:"update_time"`
	List       []hotboardItem `json:"list"`
}

type hotboardSource struct {
	Type  string
	Name  string
	Limit int
}

type hotboardResult struct {
	Source hotboardSource
	Items  []hotboardItem
}

var morningHotboardSources = []hotboardSource{
	{Type: "zhihu", Name: "知乎热榜", Limit: 10},
}

// MorningPost fetches a fresh Chinese hot-list snapshot from several industries.
func (news) MorningPost() (records []string) {
	resultChannel := make(chan hotboardResult, len(morningHotboardSources))
	var waitGroup sync.WaitGroup

	for _, source := range morningHotboardSources {
		waitGroup.Add(1)
		go func(current hotboardSource) {
			defer waitGroup.Done()
			var result hotboardResponse
			response, err := resty.New().SetTimeout(10*time.Second).R().
				SetResult(&result).
				SetQueryParam("type", current.Type).
				Get("https://uapis.cn/api/v1/misc/hotboard")
			if err != nil || response.IsError() {
				log.Printf("每日热点早报[%s]获取失败: %v", current.Name, err)
				return
			}
			resultChannel <- hotboardResult{Source: current, Items: result.List}
		}(source)
	}
	waitGroup.Wait()
	close(resultChannel)

	results := make(map[string][]hotboardItem, len(morningHotboardSources))
	for result := range resultChannel {
		results[result.Source.Type] = result.Items
	}

	seenTitles := make(map[string]struct{})
	for _, source := range morningHotboardSources {
		added := 0
		for _, item := range results[source.Type] {
			title := strings.TrimSpace(html.UnescapeString(item.Title))
			key := strings.ToLower(title)
			if title == "" {
				continue
			}
			if _, exists := seenTitles[key]; exists {
				continue
			}
			seenTitles[key] = struct{}{}
			record := title
			if link := plainHTTPURL(item.URL); link != "" {
				record += "\n🔗" + link
			}
			records = append(records, record)
			added++
			if added == source.Limit {
				break
			}
		}
	}
	return
}

func plainHTTPURL(raw string) string {
	link := strings.TrimSpace(html.UnescapeString(raw))
	if link == "" || strings.ContainsAny(link, "\r\n") {
		return ""
	}
	parsed, err := url.Parse(link)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return link
}

// GetHotTop keeps the existing optional Baidu hot-list integration.
func (news) GetHotTop() (records []dto.HotTopDataItem) {
	var respData dto.HotTop
	res := resty.New()
	resp, err := res.R().
		SetHeader("Content-Type", "application/json;chartset=utf-8").
		SetResult(&respData).
		Get("https://api.vvhan.com/api/hotlist/baiduRD")
	if err != nil {
		log.Printf("百度热榜获取失败: %s", err.Error())
		return
	}
	log.Printf("百度热榜获取结果: %s", unicodeToText(resp.String()))
	if !respData.Success {
		log.Println("百度热榜获取失败")
		return
	}
	return respData.Data
}

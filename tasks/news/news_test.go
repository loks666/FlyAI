package news

import (
	"testing"
	"time"
)

func TestFormatMorningPost(t *testing.T) {
	now := time.Date(2026, 8, 24, 8, 0, 0, 0, time.Local)
	items := []string{
		"12306回应无座票二等座同价\n🔗https://s.weibo.com/weibo?q=%2312306%23",
		"四川地震局已终止预警第三方授权\n🔗https://s.weibo.com/weibo?q=%23%E5%9B%9B%E5%B7%9D%23",
		"浙江人的餐桌已经被梭子蟹支配了\n🔗https://www.douyin.com/search/test?type=general",
		"四川宜宾发生4.7级地震\n🔗https://www.zhihu.com/search?q=Zhihu+CLI%E4%B8%8A%E7%BA%BF&search_source=Trending&utm_content=search_hot",
	}
	want := "#📅2026-08-24日实时热点早报\n\n" +
		"1. 12306回应无座票二等座同价\n🔗https://s.weibo.com/weibo?q=%2312306%23\n" +
		"2. 四川地震局已终止预警第三方授权\n🔗https://s.weibo.com/weibo?q=%23%E5%9B%9B%E5%B7%9D%23\n" +
		"3. 浙江人的餐桌已经被梭子蟹支配了\n🔗https://www.douyin.com/search/test?type=general\n" +
		"4. 四川宜宾发生4.7级地震\n🔗https://www.zhihu.com/search?q=Zhihu+CLI%E4%B8%8A%E7%BA%BF&search_source=Trending&utm_content=search_hot"
	if got := formatMorningPost(items, now); got != want {
		t.Fatalf("morning post=\n%s\nwant=\n%s", got, want)
	}
}

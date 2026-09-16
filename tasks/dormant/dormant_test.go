package dormant

import (
	"fmt"
	"github.com/loks666/FlyAI/service"
	"strings"
	"testing"
	"time"
)

func TestFormatReport(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.Local)
	members := []service.DormantMember{
		{Wxid: "wxid_1", GroupNickname: "群昵称", WechatNickname: "微信昵称1", Count: 0},
		{Wxid: "wxid_2", WechatNickname: "微信昵称2", Count: 19},
	}
	got := formatReport(now, 79, 20, members)
	for _, want := range []string{"#潜水成员统计", "📅 截至 2026-08-24", "👥 当前群成员: 79 人", "💤 本机可读取历史中发言少于20条: 2 人", "📋 仅展示前2人昵称：", "1. 群昵称", "2. 微信昵称2"} {
		if !strings.Contains(got, want) {
			t.Fatalf("report missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "微信昵称1") || strings.Contains(got, "（0条）") {
		t.Fatalf("report should prefer group nickname and omit message counts:\n%s", got)
	}
}

func TestFormatReportShowsOnlyFirstTwentyMembersByDefault(t *testing.T) {
	members := make([]service.DormantMember, 25)
	for index := range members {
		members[index].GroupNickname = fmt.Sprintf("成员-%02d", index+1)
	}
	got := formatReport(time.Now(), 100, 1, members)
	if !strings.Contains(got, "📋 仅展示前20人昵称：") || !strings.Contains(got, "20. 成员-20") {
		t.Fatalf("report did not include the first twenty members:\n%s", got)
	}
	if strings.Contains(got, "成员-21") {
		t.Fatalf("report included a member after the twentieth:\n%s", got)
	}
}

func TestFormatReportFallsBackToTenMembersOverTwoThousandRunes(t *testing.T) {
	members := make([]service.DormantMember, 20)
	for index := range members {
		members[index].GroupNickname = fmt.Sprintf("[成员-%02d]%s", index+1, strings.Repeat("长", 110))
	}
	got := formatReport(time.Now(), 100, 1, members)
	if !strings.Contains(got, "📋 仅展示前10人昵称：") || !strings.Contains(got, "10. [成员-10]") {
		t.Fatalf("oversized report did not fall back to ten members:\n%s", got)
	}
	if strings.Contains(got, "[成员-11]") {
		t.Fatalf("oversized report included a member after the tenth:\n%s", got)
	}
}

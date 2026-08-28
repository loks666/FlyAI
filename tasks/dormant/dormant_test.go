package dormant

import (
	"github.com/loks666/FlyAI/service"
	"strings"
	"testing"
	"time"
)

func TestFormatReport(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.Local)
	members := []service.DormantMember{{Wxid: "wxid_1", Nickname: "群昵称", Count: 0}, {Wxid: "wxid_2", Nickname: "个人昵称", Count: 19}}
	got := formatReport(now, 79, 20, members)
	for _, want := range []string{"#潜水成员统计", "📅 截至 2026-08-24", "👥 当前群成员: 79 人", "💤 从入群起发言少于20条: 2 人", "群昵称（0条）、个人昵称（19条）"} {
		if !strings.Contains(got, want) {
			t.Fatalf("report missing %q:\n%s", want, got)
		}
	}
}

func TestRequiresTextAttachmentUsesConservativeRuneLimit(t *testing.T) {
	if requiresTextAttachment(strings.Repeat("中", maxInlineReportRunes)) {
		t.Fatal("message at the inline limit should remain a normal text message")
	}
	if !requiresTextAttachment(strings.Repeat("中", maxInlineReportRunes+1)) {
		t.Fatal("message above the inline limit should use a TXT attachment")
	}
}

func TestFormatTextReportIncludesMemberIdentityFields(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 30, 0, 0, time.Local)
	members := []service.DormantMember{{
		Wxid:           "wxid_example",
		GroupNickname:  "群内称呼",
		WechatNickname: "微信昵称",
		WechatAccount:  "wechat_account",
		Count:          3,
	}}
	got := formatTextReport(now, 100, 20, members)
	for _, want := range []string{
		"FlyAI 潜水成员名单",
		"序号\t群昵称\t微信昵称\t微信号\twxid\t发言数",
		"1\t群内称呼\t微信昵称\twechat_account\twxid_example\t3",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("text report missing %q:\n%s", want, got)
		}
	}
}

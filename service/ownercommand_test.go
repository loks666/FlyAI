package service

import (
	"reflect"
	"testing"
)

func TestParseOwnerCommand(t *testing.T) {
	tests := []struct {
		content string
		want    string
		ok      bool
	}{
		{"@Fly AI /昨排", "", false},
		{"@Fly AI /排行榜", OwnerCommandWeekly, true},
		{"@Fly AI 排行榜", OwnerCommandWeekly, true},
		{"@Fly AI /周排", OwnerCommandWeekly, true},
		{"@Fly AI 周排", OwnerCommandWeekly, true},
		{"@Fly AI /月排", "", false},
		{"@Fly AI /年排", OwnerCommandYearly, true},
		{"@Fly AI 年排", OwnerCommandYearly, true},
		{"@Fly AI\u2005/潜水", OwnerCommandDormant, true},
		{"@Fly AI\u2005潜水", OwnerCommandDormant, true},
		{"  @Fly AI /热点  ", OwnerCommandNews, true},
		{"  @Fly AI 热点  ", OwnerCommandNews, true},
		{"@Fly AI 先发 /热点", OwnerCommandNews, true},
		{"@Fly AI /排行榜 现在", OwnerCommandWeekly, true},
		{"/排行榜", "", false},
		{"@Fly AI /日报", "", false},
		{"@Fly AI /周报", "", false},
		{"@Fly AI /未知", "", false},
		{"@Fly AI2 /排行榜", "", false},
	}
	for _, test := range tests {
		got, ok := ParseOwnerCommand(test.content, "Fly AI")
		if got != test.want || ok != test.ok {
			t.Fatalf("ParseOwnerCommand(%q)=(%q,%v), want (%q,%v)", test.content, got, ok, test.want, test.ok)
		}
	}
}

func TestParseOwnerAttackTargets(t *testing.T) {
	tests := []struct {
		content string
		want    []string
		mode    OwnerAttackMode
		ok      bool
	}{
		{"\u2005@Fly AI\u2005给我骂他\u2005@幸运黑锦鲤\u2005", []string{"幸运黑锦鲤"}, OwnerAttackChinese, true},
		{"@Fly AI给我骂他@张三", []string{"张三"}, OwnerAttackChinese, true},
		{"@@Fly AI 给我骂他@张三", []string{"张三"}, OwnerAttackChinese, true},
		{"给我骂他@张三", []string{"张三"}, OwnerAttackChinese, true},
		{"给我骂他 @昵称1，@昵称2", []string{"昵称1", "昵称2"}, OwnerAttackChinese, true},
		{"给我骂他.@昵称1，@昵称2", []string{"昵称1", "昵称2"}, OwnerAttackMorse, true},
		{"给我骂他 . @昵称1", []string{"昵称1"}, OwnerAttackMorse, true},
		{"给我骂他@幸运黑锦鲤\u2005@我要发财🦆(-48w，努力戒色)\u2005\u2005@北海（-109.5w）\u2005\u2005@小刘\u2005", []string{"幸运黑锦鲤", "我要发财🦆(-48w，努力戒色)", "北海（-109.5w）", "小刘"}, OwnerAttackChinese, true},
		{"给我骂他@昵称1，@昵称1", []string{"昵称1"}, OwnerAttackChinese, true},
		{"@Fly AI 给我骂他张三", nil, "", false},
		{"@Fly AI /儒雅随和", nil, "", false},
	}
	for _, test := range tests {
		got, mode, ok := ParseOwnerAttackTargets(test.content, "Fly AI")
		if !reflect.DeepEqual(got, test.want) || mode != test.mode || ok != test.ok {
			t.Fatalf("ParseOwnerAttackTargets(%q)=(%q,%q,%v), want (%q,%q,%v)", test.content, got, mode, ok, test.want, test.mode, test.ok)
		}
	}
}

func TestExtractRobotCommand(t *testing.T) {
	tests := []struct {
		content string
		want    string
		ok      bool
	}{
		{"@Fly AI\u2005/排行榜", "/排行榜", true},
		{"@Fly AI\u2005排行榜", "排行榜", true},
		{"@Fly AI 你好 /潜水", "/潜水", true},
		{"@Fly AI 潜水 现在", "潜水", true},
		{"@Fly AI /未知 参数", "/未知", true},
		{"@Fly AI 你好", "", false},
		{"@Fly AI 你好 潜水", "", false},
		{"你好 @Fly AI /热点", "", false},
		{"@Fly AI2 /热点", "", false},
	}
	for _, test := range tests {
		got, ok := ExtractRobotCommand(test.content, "Fly AI")
		if got != test.want || ok != test.ok {
			t.Fatalf("ExtractRobotCommand(%q)=(%q,%v), want (%q,%v)", test.content, got, ok, test.want, test.ok)
		}
	}
}

func TestIsRobotMention(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"@Fly AI", true},
		{"@Fly AI 你好", true},
		{" @Fly AI\u2005你好 ", true},
		{"@Fly AI2 你好", false},
		{"你好 @Fly AI", false},
		{"Fly AI", false},
	}
	for _, test := range tests {
		if got := IsRobotMention(test.content, "Fly AI"); got != test.want {
			t.Fatalf("IsRobotMention(%q)=%v, want %v", test.content, got, test.want)
		}
	}
}

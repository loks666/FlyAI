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

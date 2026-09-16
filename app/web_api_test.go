package app

import (
	"github.com/loks666/FlyAI/common/types"
	"github.com/loks666/FlyAI/model/vo"
	"testing"
	"time"
)

func TestActiveWebContactsFiltersInactiveAndMapsFeatures(t *testing.T) {
	now := types.DateTime(time.Date(2026, 9, 1, 9, 30, 0, 0, time.Local))
	items := []vo.FriendItem{
		{Wxid: "active", Nickname: "群聊", IsOk: true, LastActive: now, EnableWeeklyRank: true},
		{Wxid: "inactive", IsOk: false},
	}

	result := activeWebContacts(items)
	if len(result) != 1 {
		t.Fatalf("expected one active contact, got %d", len(result))
	}
	if result[0].Wxid != "active" || !result[0].EnableWeeklyRank || result[0].LastActive != "2026-09-01 09:30:00" {
		t.Fatalf("unexpected contact mapping: %#v", result[0])
	}
}

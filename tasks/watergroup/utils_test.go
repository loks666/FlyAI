package watergroup

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRollingWeekRange(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	tests := []struct {
		name  string
		now   time.Time
		start string
		end   string
	}{
		{"monday", time.Date(2026, 8, 24, 11, 0, 0, 0, location), "2026-08-17 11:00", "2026-08-24 11:00"},
		{"sunday", time.Date(2026, 8, 30, 23, 59, 0, 0, location), "2026-08-23 23:59", "2026-08-30 23:59"},
	}

	oldLocal := time.Local
	time.Local = location
	defer func() { time.Local = oldLocal }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := rollingWeekRange(tt.now)
			if got := start.Format("2006-01-02 15:04"); got != tt.start {
				t.Fatalf("start=%s, want %s", got, tt.start)
			}
			if got := end.Format("2006-01-02 15:04"); got != tt.end {
				t.Fatalf("end=%s, want %s", got, tt.end)
			}
		})
	}
}

func TestApplyRankDateFilterIncludesRollingWeekBounds(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	location := time.FixedZone("UTC+8", 8*60*60)
	now := time.Date(2026, 8, 24, 11, 58, 0, 0, location)
	tx := applyRankDateFilter(db.Table("t_message AS tm"), "week", now).Find(&[]rankUser{})
	if tx.Error != nil {
		t.Fatalf("build weekly query: %v", tx.Error)
	}
	if sql := tx.Statement.SQL.String(); !strings.Contains(sql, "tm.create_at >= ? AND tm.create_at < ?") {
		t.Fatalf("weekly SQL is missing rolling bounds: %s", sql)
	}
	if len(tx.Statement.Vars) != 2 {
		t.Fatalf("weekly query has %d bound values, want 2", len(tx.Statement.Vars))
	}
	if got := tx.Statement.Vars[0].(time.Time); !got.Equal(now.AddDate(0, 0, -7)) {
		t.Fatalf("start=%s, want %s", got, now.AddDate(0, 0, -7))
	}
	if got := tx.Statement.Vars[1].(time.Time); !got.Equal(now) {
		t.Fatalf("end=%s, want %s", got, now)
	}
}

func TestExcludeRankUserFiltersRobotWxid(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	tx := excludeRankUser(db.Table("t_message AS tm"), "wxid_fly_ai").Find(&[]rankUser{})
	if tx.Error != nil {
		t.Fatalf("build robot exclusion query: %v", tx.Error)
	}
	if sql := tx.Statement.SQL.String(); !strings.Contains(sql, "tm.group_user <> ?") {
		t.Fatalf("ranking SQL is missing robot exclusion: %s", sql)
	}
	if len(tx.Statement.Vars) != 1 || tx.Statement.Vars[0] != "wxid_fly_ai" {
		t.Fatalf("ranking query vars=%v, want robot wxid", tx.Statement.Vars)
	}
}

func TestFormatMessageBreakdown(t *testing.T) {
	record := rankUser{TextCount: 12, VideoCount: 2, EmoticonCount: 3, ShareCount: 4}
	want := "文本12, 视频2, 表情3, 转发4"
	if got := formatMessageBreakdown(record); got != want {
		t.Fatalf("breakdown=%q, want %q", got, want)
	}
}

func TestFormatRankingBreakdown(t *testing.T) {
	record := rankUser{ImageCount: 2, VideoCount: 2, ShareCount: 1}
	want := "🖼️2 , 🎬2 , 🔁1"
	if got := formatRankingBreakdown(record); got != want {
		t.Fatalf("ranking breakdown=%q, want %q", got, want)
	}
}

func TestRankBadgeUsesNumericKeycapsForFourthThroughTenth(t *testing.T) {
	want := []string{"🥇", "🥈", "🥉", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	for index, expected := range want {
		if got := rankBadge(index); got != expected {
			t.Fatalf("rankBadge(%d)=%q, want %q", index, got, expected)
		}
	}
	if got := rankBadge(10); got != "🏆" {
		t.Fatalf("rankBadge(10)=%q, want fallback trophy", got)
	}
}

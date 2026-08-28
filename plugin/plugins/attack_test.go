package plugins

import (
	"github.com/loks666/FlyAI/service"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRandomAttackReplyLoadsJSONList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attack.json")
	if err := os.WriteFile(path, []byte(`["", " first ", "second"]`), 0o600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		reply, err := randomAttackReply(path)
		if err != nil {
			t.Fatal(err)
		}
		if reply != "first" && reply != "second" {
			t.Fatalf("unexpected reply %q", reply)
		}
	}
}

func TestRandomAttackReplyRejectsEmptyList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attack.json")
	if err := os.WriteFile(path, []byte(`[" "]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := randomAttackReply(path); err == nil {
		t.Fatal("expected an error for an empty reply list")
	}
}

func TestAttackReplyListsSelectChineseAndMorseSeparately(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attack.json")
	if err := os.WriteFile(path, []byte(`{"chinese":["中文"],"morse":["-. .."]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	chinese, err := loadAttackReplies(path, service.OwnerAttackChinese)
	if err != nil || len(chinese) != 1 || chinese[0] != "中文" {
		t.Fatalf("unexpected chinese list: %v, %v", chinese, err)
	}
	morse, err := loadAttackReplies(path, service.OwnerAttackMorse)
	if err != nil || len(morse) != 1 || morse[0] != "-. .." {
		t.Fatalf("unexpected morse list: %v, %v", morse, err)
	}
}

func TestNonOwnerMentionReplyUsesPoolWhenWindowIsAvailable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attack.json")
	if err := os.WriteFile(path, []byte(`["random reply"]`), 0o600); err != nil {
		t.Fatal(err)
	}
	reply, err := nonOwnerMentionReply(service.NonOwnerReplyRandom, path)
	if err != nil {
		t.Fatal(err)
	}
	if reply != "random reply" {
		t.Fatalf("reply = %q, want random reply", reply)
	}
}

func TestAdministratorRechargeReplyFormat(t *testing.T) {
	if got := formatAdministratorRechargeReply("管理员", "admin_account"); got != "您的余额不足，请联系管理员【管理员，admin_account】充值，未来一小时内不再回复您的消息" {
		t.Fatalf("unexpected reply: %q", got)
	}
}

func TestFormatTargetAttackReplyUsesPlainMentionText(t *testing.T) {
	if got := formatTargetAttackReply(" 目标昵称 ", " 随机一句 "); got != "@目标昵称 随机一句" {
		t.Fatalf("unexpected target attack reply: %q", got)
	}
}

func TestMultipleTargetAttackReplyUsesOneLinePerTarget(t *testing.T) {
	lines := []string{
		formatTargetAttackReply("昵称1", "第一句"),
		formatTargetAttackReply("昵称2", "第二句"),
	}
	if got := strings.Join(lines, "\n"); got != "@昵称1 第一句\n@昵称2 第二句" {
		t.Fatalf("unexpected multiple target reply: %q", got)
	}
}

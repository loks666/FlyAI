package service

import "testing"

func TestNormalizeRobotWechatAccount(t *testing.T) {
	account, err := normalizeRobotWechatAccount("  niubaidely  ")
	if err != nil || account != "niubaidely" {
		t.Fatalf("unexpected normalized account %q, err=%v", account, err)
	}
	if _, err := normalizeRobotWechatAccount("two accounts"); err == nil {
		t.Fatal("account containing spaces must be rejected")
	}
}
